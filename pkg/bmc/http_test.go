package bmc

import (
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

func getRoot() []byte {
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

func getBrokenRoot() []byte {
	return []byte(`{
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		}
	}`)
}

func getBrokenRoot2() []byte {
	return []byte(`{
		"AccountService": {
			"@odata.id": "/redfish/v1/AccountService"
		},
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		}
	}`)
}

func Test_patchMissingSessionsLink_AddsSessionsLink(t *testing.T) {
	t.Parallel()

	orig := slices.Clone(getBrokenRoot2())
	out, err := patchMissingSessionsLink(orig)
	require.NoError(t, err)

	var outRoot map[string]any
	err = json.Unmarshal(out, &outRoot)
	require.NoError(t, err)

	// check added Sessions link
	links, ok := outRoot["Links"].(map[string]any)
	require.True(t, ok)
	sessions, ok := links["Sessions"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "/redfish/v1/SessionService/Sessions", sessions["@odata.id"])

	// no change to rest of document
	accountService, ok := outRoot["AccountService"].(map[string]any)
	require.True(t, ok)
	accountServiceID, ok := accountService["@odata.id"]
	require.True(t, ok)
	assert.Equal(t, "/redfish/v1/AccountService", accountServiceID)

	sessionService, ok := outRoot["SessionService"].(map[string]any)
	require.True(t, ok)
	sessionServiceID, ok := sessionService["@odata.id"]
	require.True(t, ok)
	assert.Equal(t, "/redfish/v1/SessionService", sessionServiceID)
}

func Test_patchMissingSessionsLink_NoChangeIfLinksExist(t *testing.T) {
  t.Parallel()

	orig := slices.Clone(getRoot())
	out, err := patchMissingSessionsLink(orig)
	require.NoError(t, err)
	assert.JSONEq(t, string(getRoot()), string(out))
}

func Test_patchMissingSessionsLink_InvalidJSON(t *testing.T) {
  t.Parallel()

	orig := []byte(`{invalid json}`)
	out, err := patchMissingSessionsLink(orig)
	require.Error(t, err)
	assert.Equal(t, orig, out)
}

func Test_modifierTransport_ModifiesResponse(t *testing.T) {
  t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redfish/v1/" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json;charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(getBrokenRoot())
			assert.NoError(t, err)

			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := newHTTPClient(true, nil)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/redfish/v1/", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	out, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(out), string(getRoot()))
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

func Test_NewHttpClient_WithDialer(t *testing.T) {
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
