package cli

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLFlag_String_Nil(t *testing.T) {
  t.Parallel()

	var flag *URLFlag

	assert.Empty(t, flag.String())
}

func TestURLFlag_String_Empty(t *testing.T) {
  t.Parallel()

	flag := &URLFlag{}
	assert.Empty(t, flag.String())
}

func TestURLFlag_String_HTTP(t *testing.T) {
  t.Parallel()

	flag := &URLFlag{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/path",
	}
	assert.Equal(t, "http://example.com/path", flag.String())
}

func TestURLFlag_String_HTTPS_WithPort(t *testing.T) {
  t.Parallel()

	flag := &URLFlag{
		Scheme: "https",
		Host:   "example.com:8080",
		Path:   "/api/v1",
	}
	assert.Equal(t, "https://example.com:8080/api/v1", flag.String())
}

func TestURLFlag_String_WithQuery(t *testing.T) {
  t.Parallel()

	flag := &URLFlag{
		Scheme:   "https",
		Host:     "api.example.com",
		Path:     "/search",
		RawQuery: "q=test&limit=10",
	}
	assert.Equal(t, "https://api.example.com/search?q=test&limit=10", flag.String())
}

func TestURLFlag_Set_ValidHTTP(t *testing.T) {
  t.Parallel()

	var flag URLFlag

	err := flag.Set("http://example.com")
	require.NoError(t, err)
	assertURLComponents(t, flag, "http", "example.com", "", "")
}

func TestURLFlag_Set_ValidHTTPS_WithPath(t *testing.T) {
  t.Parallel()

	var flag URLFlag

	err := flag.Set("https://api.example.com/v1/users")
	require.NoError(t, err)
	assertURLComponents(t, flag, "https", "api.example.com", "/v1/users", "")
}

func TestURLFlag_Set_WithPortAndQuery(t *testing.T) {
  t.Parallel()

	var flag URLFlag

	err := flag.Set("https://localhost:8080/api?version=1")
	require.NoError(t, err)
	assertURLComponents(t, flag, "https", "localhost:8080", "/api", "version=1")
}

func TestURLFlag_Set_RelativePath(t *testing.T) {
  t.Parallel()

	var flag URLFlag

	err := flag.Set("/api/v1")
	require.NoError(t, err)
	assertURLComponents(t, flag, "", "", "/api/v1", "")
}

func TestURLFlag_Set_InvalidURL(t *testing.T) {
  t.Parallel()

	var flag URLFlag

	err := flag.Set("http://example.com/\x00")
	require.Error(t, err)
}

func TestURLFlag_Type(t *testing.T) {
  t.Parallel()

	flag := &URLFlag{}
	assert.Equal(t, "url", flag.Type())
}

func TestURLFlag_RoundTrip_HTTP(t *testing.T) {
  t.Parallel()
	assertRoundTrip(t, "http://example.com")
}

func TestURLFlag_RoundTrip_HTTPS_Complex(t *testing.T) {
  t.Parallel()
	assertRoundTrip(t, "https://api.example.com:8080/v1/users?limit=10")
}

func TestURLFlag_RoundTrip_FTP(t *testing.T) {
  t.Parallel()
	assertRoundTrip(t, "ftp://files.example.com/path/to/file.txt")
}

func TestURLFlag_RoundTrip_RelativePath(t *testing.T) {
  t.Parallel()
	assertRoundTrip(t, "/relative/path")
}

func TestURLFlag_RoundTrip_HostPath(t *testing.T) {
  t.Parallel()
	assertRoundTrip(t, "example.com/path")
}

// Helper function to reduce duplication in round-trip tests.
func assertRoundTrip(t *testing.T, original string) {
	t.Helper()

	var flag URLFlag

	err := flag.Set(original)
	require.NoError(t, err)

	result := flag.String()

	// Parse both to compare normalized forms
	originalURL, err := url.Parse(original)
	require.NoError(t, err)

	resultURL, err := url.Parse(result)
	require.NoError(t, err)

	assert.Equal(t, originalURL.String(), resultURL.String())
}

func assertURLComponents(t *testing.T, flag URLFlag,
	scheme, host, path, query string,
) {
	t.Helper()

	assert.Equal(t, flag.Scheme, scheme)
	assert.Equal(t, flag.Host, host)
	assert.Equal(t, flag.Path, path)
	assert.Equal(t, flag.RawQuery, query)
}
