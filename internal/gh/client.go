package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"encoding/base64"
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

// GHReadme represents the response for a README file from the GitHub API.
type GHReadme struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// GetReadme fetches the README content for a single repository.
func GetReadme(ctx context.Context, client *http.Client, fullName string) (string, error) {
	readmeURL := fmt.Sprintf("https://api.github.com/repos/%s/readme", fullName)
	var req *http.Request
	var err error
	req, err = http.NewRequestWithContext(ctx, "GET", readmeURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	var resp *http.Response
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		// Don't print retry message for READMEs to avoid spamming the console.
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// It's common for repos to not have a README, so we treat 404 as non-fatal.
		if resp.StatusCode == http.StatusNotFound {
			return "", nil // No README, not an error.
		}
		return "", fmt.Errorf("failed to get readme for %s: %s", fullName, resp.Status)
	}

	var readme GHReadme
	if err := json.NewDecoder(resp.Body).Decode(&readme); err != nil {
		return "", err
	}

	if readme.Encoding != "base64" {
		return "", fmt.Errorf("unknown readme encoding: %s", readme.Encoding)
	}

	decodedContent, err := base64.StdEncoding.DecodeString(readme.Content)
	if err != nil {
		return "", fmt.Errorf("could not decode readme content: %w", err)
	}

	return string(decodedContent), nil
}

// GetStarredRepos fetches starred repositories for the authenticated user, up to a given limit.
func GetStarredRepos(ctx context.Context, token, proxyAddr string, limit int) ([]GHRepo, error) {
	client, err := NewClient(proxyAddr, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	var allRepos []GHRepo
	// We can fetch up to 100 per page.
	nextURL := "https://api.github.com/user/starred?per_page=100"

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

		if limit > 0 && len(allRepos) >= limit {
			// Trim excess repos if we fetched more than the limit on the last page
			allRepos = allRepos[:limit]
			break
		}

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
