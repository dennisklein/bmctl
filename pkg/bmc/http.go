package bmc

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/proxy"
)

type (
	// modifierTransport is an http.RoundTripper that wraps another RoundTripper.
	// It allows modification of requests or responses before/after passing them to the original transport.
	modifierTransport struct {
		original http.RoundTripper
	}
)

// RoundTrip executes a single HTTP transaction, modifying the response if needed.
// It wraps the original RoundTripper and applies patchServiceRoot to the response.
func (t *modifierTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.original.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if err := patchServiceRoot(req, resp); err != nil {
		// If patching fails, close the response body to avoid leaks
		if cerr := resp.Body.Close(); cerr != nil {
			return nil, errors.Join(err, cerr)
		}

		return nil, err
	}

	return resp, nil
}

// patchMissingSessionsLink checks if the ServiceRoot JSON body is missing the required
// Sessions Link (common in some older Huawei iBMC versions) and adds it if necessary.
// It returns the possibly modified body and any error encountered.
func patchMissingSessionsLink(body []byte) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return body, err
	}

	// If "Links" already exists, nothing to patch
	_, ok := root["Links"]
	if ok {
		return body, nil
	}

	// Add the missing Sessions link
	if sessionService, ok := root["SessionService"].(map[string]any); ok {
		if sessionServiceID, ok := sessionService["@odata.id"].(string); ok {
			root["Links"] = map[string]any{
				"Sessions": map[string]any{
					"@odata.id": sessionServiceID + "/Sessions",
				},
			}
		}
	}

	return json.Marshal(root)
}

// patchServiceRoot modifies the HTTP response for the Redfish ServiceRoot endpoint ("/redfish/v1/").
// It reads and potentially patches the response body to add a missing Sessions link (for compatibility with some BMCs),
// then updates the response body and headers. Returns an error if reading or patching the body fails.
func patchServiceRoot(req *http.Request, resp *http.Response) error {
	if !strings.HasSuffix(req.URL.Path, "/redfish/v1/") {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	cerr := resp.Body.Close()

	if cerr != nil || err != nil {
		return errors.Join(err, cerr)
	}

	newBody, err := patchMissingSessionsLink(body)
	if err != nil {
		return err
	}

	if !bytes.Equal(body, newBody) {
		resp.Body = io.NopCloser(bytes.NewReader(newBody))
		resp.ContentLength = int64(len(newBody))
		resp.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
	} else {
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	return nil
}

// NewHttpClient creates and returns a new *http.Client.
// If the 'insecure' parameter is true, the client will skip TLS certificate verification.
// The 'dialer' parameter allows specifying a custom proxy dialer for the HTTP transport.
// Returns the configured *http.Client or an error if the default transport cannot be obtained.
func newHTTPClient(insecure bool, dialer proxy.Dialer) (*http.Client, error) {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("could not get http.DefaultTransport")
	}
	transport := defaultTransport.Clone()
	if dialer != nil {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	}
	if insecure {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}
	return &http.Client{Transport: &modifierTransport{original: transport}}, nil
}
