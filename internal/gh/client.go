package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// NewClient creates a new HTTP client, optionally configured with a proxy and auth token.
func NewClient(proxyAddr, token string) (*http.Client, error) {
	transport := &http.Transport{}

	if proxyAddr != "" {
		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
	}

	if token != "" {
		// Add a wrapper to the transport to inject the auth header.
		client.Transport = &authTransport{
			token:     token,
			transport: transport,
		}
	}

	return client, nil
}

// authTransport is a wrapper to add the Authorization header to requests.
type authTransport struct {
	token     string
	transport http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.transport.RoundTrip(req)
}

// GHRepo represents a repository as returned by the GitHub API.
type GHRepo struct {
	ID              int64  `json:"id"`
	FullName        string `json:"full_name"`
	Description     string `json:"description"`
	HTMLURL         string `json:"html_url"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
}

// GetStarredRepos fetches all starred repositories for the authenticated user.
func GetStarredRepos(ctx context.Context, token, proxyAddr string) ([]GHRepo, error) {
	client, err := NewClient(proxyAddr, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	var allRepos []GHRepo
	nextURL := "https://api.github.com/user/starred"

	for nextURL != "" {
		var req *http.Request
		var err error
		req, err = http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		var resp *http.Response
		const maxRetries = 3
		for i := 0; i < maxRetries; i++ {
			resp, err = client.Do(req)
			if err == nil {
				break
			}
			fmt.Printf("Request failed (attempt %d/%d): %v. Retrying in 2 seconds...\n", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("github api returned non-200 status: %s", resp.Status)
		}

		var repos []GHRepo
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		allRepos = append(allRepos, repos...)

		// Handle pagination
		linkHeader := resp.Header.Get("Link")
		nextURL = parseNextLink(linkHeader)
	}

	return allRepos, nil
}

// parseNextLink extracts the next page URL from the Link header.
func parseNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	links := strings.Split(linkHeader, ",")
	re := regexp.MustCompile(`<(.+?)>; rel="next"`)
	for _, link := range links {
		match := re.FindStringSubmatch(link)
		if len(match) > 1 {
			return match[1]
		}
	}
	return ""
}
