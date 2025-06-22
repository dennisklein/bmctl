package cli

import "net/url"

// URLFlag is a strongly typed alias for url.URL that implements pflag.Value.
type URLFlag url.URL

// String returns the URL as a string, implementing part of pflag.Value.
func (u *URLFlag) String() string {
	if u == nil {
		return ""
	}

	underlying := (*url.URL)(u)

	return underlying.String()
}

// Set parses a string into a URL, implementing part of pflag.Value.
func (u *URLFlag) Set(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	*u = URLFlag(*parsedURL)

	return nil
}

// Type returns the type name for this flag value, implementing part of pflag.Value.
func (u *URLFlag) Type() string {
	return "url"
}
