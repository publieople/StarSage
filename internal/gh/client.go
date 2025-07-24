package gh

import (
	"net/http"
	"net/url"
)

// NewClient creates a new HTTP client, optionally configured with a proxy.
func NewClient(proxyAddr string) (*http.Client, error) {
	transport := &http.Transport{}

	if proxyAddr != "" {
		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: transport,
	}, nil
}
