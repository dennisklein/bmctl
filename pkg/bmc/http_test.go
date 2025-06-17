// SPDX-FileCopyrightText: 2025 GSI Helmholtzzentrum für Schwerionenforschung GmbH <https://www.gsi.de/en/>
//
// SPDX-License-Identifier: LGPL-3.0-or-later

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

var (
	root = []byte(`{
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		},
		"Links": {
			"Sessions": {
				"@odata.id": "/redfish/v1/SessionService/Sessions"
			}
		}
	}`)
	brokenRoot = []byte(`{
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		}
	}`)
	brokenRoot2 = []byte(`{
		"AccountService": {
			"@odata.id": "/redfish/v1/AccountService"
		},
		"SessionService": {
			"@odata.id": "/redfish/v1/SessionService"
		}
	}`)
)

func Test_patchMissingSessionsLink_AddsSessionsLink(t *testing.T) {
	orig := slices.Clone(brokenRoot2)
	out, err := patchMissingSessionsLink(orig)
	require.NoError(t, err)

	var outRoot map[string]any
	err = json.Unmarshal(out, &outRoot)
	require.NoError(t, err)

	// check added Sessions link
	links := outRoot["Links"].(map[string]any)
	sessions := links["Sessions"].(map[string]any)
	assert.Equal(t, "/redfish/v1/SessionService/Sessions", sessions["@odata.id"])

	// no change to rest of document
	assert.Equal(
		t,
		"/redfish/v1/AccountService",
		outRoot["AccountService"].(map[string]any)["@odata.id"],
	)
	assert.Equal(
		t,
		"/redfish/v1/SessionService",
		outRoot["SessionService"].(map[string]any)["@odata.id"],
	)
}

func Test_patchMissingSessionsLink_NoChangeIfLinksExist(t *testing.T) {
	orig := slices.Clone(root)
	out, err := patchMissingSessionsLink(orig)
	require.NoError(t, err)
	assert.JSONEq(t, string(root), string(out))
}

func Test_patchMissingSessionsLink_InvalidJSON(t *testing.T) {
	orig := []byte(`{invalid json}`)
	out, err := patchMissingSessionsLink(orig)
	require.Error(t, err)
	assert.Equal(t, orig, out)
}

func Test_modifierTransport_ModifiesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redfish/v1/" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json;charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(brokenRoot)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := newHttpClient(true, nil)
	require.NoError(t, err)

	resp, err := client.Get(server.URL + "/redfish/v1/")
	require.NoError(t, err)
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, string(out), string(root))
}

func Test_NewHttpClient_Insecure(t *testing.T) {
	client, err := newHttpClient(true, nil)
	require.NoError(t, err)
	tr := client.Transport.(*modifierTransport)
	underlying := tr.original.(*http.Transport)
	require.NotNil(t, underlying.TLSClientConfig)
	assert.True(t, underlying.TLSClientConfig.InsecureSkipVerify)
}

// mockDialer implements proxy.Dialer for testing
type mockDialer struct {
	Called bool
}

func (m *mockDialer) Dial(network, addr string) (net.Conn, error) {
	m.Called = true
	return nil, nil
}

func Test_NewHttpClient_WithDialer(t *testing.T) {
	dialer := &mockDialer{}
	client, err := newHttpClient(false, dialer)
	require.NoError(t, err)
	transport, ok := client.Transport.(*modifierTransport)
	require.True(t, ok)
	origTransport, ok := transport.original.(*http.Transport)
	if ok && origTransport.DialContext != nil {
		origTransport.DialContext(nil, "tcp", "example.com:80")
	}
	assert.True(t, dialer.Called, "expected dialer to be called")
}
