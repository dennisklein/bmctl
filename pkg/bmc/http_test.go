package bmc

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions for creating different JSON payloads

func getValidServiceRoot() []byte {
	return []byte(`{
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		},
		"Links": {
			"Sessions": {
				"@odata.id": "/redfish/v1/SessionService/Sessions"
			}
		}
	}`)
}

func getServiceRootMissingSessions() []byte {
	return []byte(`{
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		}
	}`)
}

func getServiceRootWithMultipleServices() []byte {
	return []byte(`{
		"AccountService": {
			"@odata.id": "/redfish/v1/AccountService"
		},
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		}
	}`)
}

func getInvalidJSON() []byte {
	return []byte(`{invalid json}`)
}

// Helper function to create test HTTP server.
func createTestServer(responseBody []byte, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redfish/v1/" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json;charset=utf-8")
			w.WriteHeader(statusCode)

			_, err := w.Write(responseBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		http.NotFound(w, r)
	}))
}

// Helper function to parse JSON and extract nested field.
func extractSessionsLink(data []byte) (string, error) {
	var root map[string]any

	err := json.Unmarshal(data, &root)
	if err != nil {
		return "", err
	}

	if links, ok := root["Links"].(map[string]any); ok {
		if sessions, ok := links["Sessions"].(map[string]any); ok {
			if id, ok := sessions["@odata.id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", nil
}

func Test_patchMissingSessionsLink_AddsSessionsLink(t *testing.T) {
	t.Parallel()

	orig := slices.Clone(getServiceRootWithMultipleServices())
	out, err := patchMissingSessionsLink(orig)
	require.NoError(t, err)

	// Verify Sessions link was added
	sessionsLink, err := extractSessionsLink(out)
	require.NoError(t, err)
	assert.Equal(t, "/redfish/v1/SessionService/Sessions", sessionsLink)

	// Verify other services remain unchanged
	var outRoot map[string]any

	err = json.Unmarshal(out, &outRoot)
	require.NoError(t, err)

	accountService, ok := outRoot["AccountService"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "/redfish/v1/AccountService", accountService["@odata.id"])

	sessionService, ok := outRoot["SessionService"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "/redfish/v1/SessionService", sessionService["@odata.id"])
}

func Test_patchMissingSessionsLink_NoChangeIfLinksExist(t *testing.T) {
	t.Parallel()

	orig := slices.Clone(getValidServiceRoot())
	out, err := patchMissingSessionsLink(orig)
	require.NoError(t, err)
	assert.JSONEq(t, string(getValidServiceRoot()), string(out))
}

func Test_patchMissingSessionsLink_InvalidJSON(t *testing.T) {
	t.Parallel()

	orig := getInvalidJSON()
	out, err := patchMissingSessionsLink(orig)
	require.Error(t, err)
	assert.Equal(t, orig, out)
}

func Test_modifierTransport_ModifiesResponse(t *testing.T) {
	t.Parallel()

	server := createTestServer(getServiceRootMissingSessions(), http.StatusOK)
	defer server.Close()

	client, err := newHTTPClient(true, nil)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		server.URL+"/redfish/v1/",
		nil,
	)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	out, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Verify that Sessions link was added
	sessionsLink, err := extractSessionsLink(out)
	require.NoError(t, err)
	assert.Equal(t, "/redfish/v1/SessionService/Sessions", sessionsLink)
}

func Test_newHTTPClient_Insecure(t *testing.T) {
	t.Parallel()

	client, err := newHTTPClient(true, nil)
	require.NoError(t, err)

	tr, ok := client.Transport.(*modifierTransport)
	require.True(t, ok)

	underlying, ok := tr.original.(*http.Transport)
	require.True(t, ok)

	require.NotNil(t, underlying.TLSClientConfig)
	assert.True(t, underlying.TLSClientConfig.InsecureSkipVerify)
}

// mockDialer implements proxy.Dialer for testing.
type mockDialer struct {
	Called bool
}

func (m *mockDialer) Dial(network, addr string) (net.Conn, error) {
	m.Called = true
	return nil, nil //nolint
}

func Test_NewHTTPClient_WithDialer(t *testing.T) {
	t.Parallel()

	dialer := &mockDialer{}
	client, err := newHTTPClient(false, dialer)
	require.NoError(t, err)

	transport, ok := client.Transport.(*modifierTransport)
	require.True(t, ok)

	origTransport, ok := transport.original.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, origTransport.DialContext)

	_, err = origTransport.DialContext(t.Context(), "tcp", "example.com:80")
	require.NoError(t, err)

	assert.True(t, dialer.Called, "expected dialer to be called")
}

func Test_newHTTPClient_SecureByDefault(t *testing.T) {
	t.Parallel()

	client, err := newHTTPClient(false, nil)
	require.NoError(t, err)

	transport, ok := client.Transport.(*modifierTransport)
	require.True(t, ok)

	origTransport, ok := transport.original.(*http.Transport)
	require.True(t, ok)

	// Should not have InsecureSkipVerify set to true
	if origTransport.TLSClientConfig != nil {
		assert.False(t, origTransport.TLSClientConfig.InsecureSkipVerify)
	}
}

func Test_patchServiceRoot_NonRedfishPath(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"https://example.com/other/path",
		nil,
	)
	require.NoError(t, err)

	resp := &http.Response{
		Body: io.NopCloser(bytes.NewReader([]byte("test content"))),
	}

	err = patchServiceRoot(req, resp)
	require.NoError(t, err)

	// Should not modify non-redfish paths
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(body))
}

func Test_patchMissingSessionsLink_NoSessionService(t *testing.T) {
	t.Parallel()

	input := []byte(`{
		"AccountService": {
			"@odata.id": "/redfish/v1/AccountService"
		}
	}`)

	out, err := patchMissingSessionsLink(input)
	require.NoError(t, err)

	// Should not add Sessions link if SessionService is missing
	assert.JSONEq(t, string(input), string(out))
}
