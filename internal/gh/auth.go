package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Constants for the device flow
const (
	clientID       = "Ov23liMsNArPWi0a8BWd" // User-provided Client ID
	deviceCodeURL  = "https://github.com/login/device/code"
	accessTokenURL = "https://github.com/login/oauth/access_token"
	grantType      = "urn:ietf:params:oauth:grant-type:device_code"
)

// DeviceFlowResponse holds the response from the device code request.
type DeviceFlowResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// AccessTokenResponse holds the response from the access token request.
type AccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// PerformDeviceFlow handles the entire GitHub OAuth Device Flow.
func PerformDeviceFlow(ctx context.Context, proxyAddr string) (string, error) {
	client, err := NewClient(proxyAddr)
	if err != nil {
		return "", fmt.Errorf("failed to create http client: %w", err)
	}

	// Step 1: Get Device and User Codes
	deviceFlowResp, err := requestDeviceCode(ctx, client)
	if err != nil {
		return "", fmt.Errorf("failed to request device code: %w", err)
	}

	// Step 2: Display instructions to the user
	fmt.Printf("\nYour one-time code is: %s\n", deviceFlowResp.UserCode)
	fmt.Printf("Please go to %s to authorize.\n", deviceFlowResp.VerificationURI)

	// Step 3: Poll for the access token
	return pollForAccessToken(ctx, client, deviceFlowResp)
}

func requestDeviceCode(ctx context.Context, client *http.Client) (*DeviceFlowResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "repo,user")

	req, err := http.NewRequestWithContext(ctx, "POST", deviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get device code: %s", string(body))
	}

	var deviceFlowResp DeviceFlowResponse
	if err := json.Unmarshal(body, &deviceFlowResp); err != nil {
		return nil, err
	}

	return &deviceFlowResp, nil
}

func pollForAccessToken(ctx context.Context, client *http.Client, deviceFlowResp *DeviceFlowResponse) (string, error) {
	interval := time.Duration(deviceFlowResp.Interval) * time.Second
	timeout := time.After(time.Duration(deviceFlowResp.ExpiresIn) * time.Second)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Println("Waiting for authorization...")

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("authentication timed out after %d minutes", deviceFlowResp.ExpiresIn/60)
		case <-ticker.C:
			data := url.Values{}
			data.Set("client_id", clientID)
			data.Set("device_code", deviceFlowResp.DeviceCode)
			data.Set("grant_type", grantType)

			req, err := http.NewRequestWithContext(ctx, "POST", accessTokenURL, strings.NewReader(data.Encode()))
			if err != nil {
				fmt.Printf("error creating poll request: %v. Retrying...\n", err)
				continue
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("network error while polling: %v. Retrying...\n", err)
				continue
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var tokenResp AccessTokenResponse
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				fmt.Printf("error unmarshalling token response: %v. Retrying...\n", err)
				continue
			}

			switch tokenResp.Error {
			case "":
				if tokenResp.AccessToken != "" {
					return tokenResp.AccessToken, nil
				}
			case "authorization_pending":
				// Continue polling.
			case "slow_down":
				interval += 5 * time.Second
				ticker.Reset(interval)
			case "expired_token":
				return "", fmt.Errorf("the device code has expired: %s", tokenResp.ErrorDescription)
			case "access_denied":
				return "", fmt.Errorf("authentication was denied by the user: %s", tokenResp.ErrorDescription)
			default:
				return "", fmt.Errorf("unexpected error during polling: %s - %s", tokenResp.Error, tokenResp.ErrorDescription)
			}
		}
	}
}
