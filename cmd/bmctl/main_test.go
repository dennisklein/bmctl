package main

import (
	"net/url"
	"testing"

	"github.com/GSI-HPC/bmctl/pkg/bmc"
	"github.com/stretchr/testify/assert"
)

func parseURL(s string) url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return *u
}

// newTestConfig creates a basic test configuration with the given endpoint.
func newTestConfig(endpoint string) *bmc.ClientConfig {
	return &bmc.ClientConfig{
		Endpoint: parseURL(endpoint),
		SSHProxy: "",
	}
}

// longString generates a string of the specified length filled with the given character.
func longString(length int, char byte) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = char
	}

	return string(b)
}

// assertValidationError tests the validation function and asserts the expected error.
func assertValidationError(t *testing.T, cfg *bmc.ClientConfig, expectedError string) {
	t.Helper()

	err := validateBmcClientConfig(cfg)
	if expectedError == "" {
		assert.NoError(t, err)
	} else {
		assert.EqualError(t, err, expectedError)
	}
}

func Test_validateBmcClientConfig_validHttpsEndpoint(t *testing.T) {
	cfg := newTestConfig("https://bmc.example.com")
	assertValidationError(t, cfg, "")
}

func Test_validateBmcClientConfig_validHttpEndpoint(t *testing.T) {
	cfg := newTestConfig("http://bmc.example.com")
	assertValidationError(t, cfg, "")
}

func Test_validateBmcClientConfig_invalidFtpScheme(t *testing.T) {
	cfg := newTestConfig("ftp://bmc.example.com")
	assertValidationError(t, cfg, "endpoint must be a valid http(s) URL")
}

func Test_validateBmcClientConfig_invalidSshScheme(t *testing.T) {
	cfg := newTestConfig("ssh://bmc.example.com")
	assertValidationError(t, cfg, "endpoint must be a valid http(s) URL")
}

func Test_validateBmcClientConfig_emptyScheme(t *testing.T) {
	cfg := &bmc.ClientConfig{
		Endpoint: url.URL{Host: "bmc.example.com"},
		SSHProxy: "",
	}
	assertValidationError(t, cfg, "endpoint must be a valid http(s) URL")
}

func Test_validateBmcClientConfig_endpointTooLong(t *testing.T) {
	longHost := longString(250, 'a')
	cfg := newTestConfig("https://" + longHost + ".com")
	assertValidationError(t, cfg, "endpoint URL too long (max 253 characters)")
}

func Test_validateBmcClientConfig_userTooLong(t *testing.T) {
	cfg := newTestConfig("https://bmc.example.com")
	cfg.User = longString(65, 'u')
	assertValidationError(t, cfg, "user too long (max 64 characters)")
}

func Test_validateBmcClientConfig_passwordTooLong(t *testing.T) {
	cfg := newTestConfig("https://bmc.example.com")
	cfg.Password = longString(129, 'p')
	assertValidationError(t, cfg, "password too long (max 128 characters)")
}

func Test_validateBmcClientConfig_sshProxyTooLong(t *testing.T) {
	cfg := newTestConfig("https://bmc.example.com")
	cfg.SSHProxy = longString(254, 's')
	assertValidationError(t, cfg, "ssh-proxy too long (max 253 characters)")
}
