package bmc

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a basic ClientConfig for testing.
func createTestClientConfig(endpoint string) ClientConfig {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		// For test cases with invalid URLs, provide a zero URL
		parsed = &url.URL{}
	}

	return ClientConfig{
		Endpoint: *parsed,
		User:     "testuser",
		Password: "testpass",
		Insecure: true,
	}
}

func TestClientConfig_Creation(t *testing.T) {
	t.Parallel()

	cfg := createTestClientConfig("https://bmc.example.com")

	assert.Equal(t, "bmc.example.com", cfg.Endpoint.Host)
	assert.Equal(t, "testuser", cfg.User)
	assert.Equal(t, "testpass", cfg.Password)
	assert.True(t, cfg.Insecure)
}

func TestClient_CloseHandlesNilFields(t *testing.T) {
	t.Parallel()

	// Test that Close doesn't panic with nil fields
	client := &Client{}

	assert.NotPanics(t, func() {
		client.Close()
	})
}
