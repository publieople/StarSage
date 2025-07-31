package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OllamaProvider implements the Provider interface for Ollama.
type OllamaProvider struct {
	model   string
	baseURL string
	client  *http.Client
}

// NewOllamaProvider creates a new provider for Ollama.
// baseURL defaults to http://127.0.0.1:11434 if empty.
func NewOllamaProvider(model, baseURL string, client *http.Client) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	return &OllamaProvider{
		model:   model,
		baseURL: baseURL,
		client:  client,
	}
}

// ollamaGenerateRequest is the request body for the Ollama API.
type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaGenerateResponse is a single response object from the streaming API.
type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Generate sends a prompt to the Ollama API and returns the response.
func (p *OllamaProvider) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody, err := json.Marshal(ollamaGenerateRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: true, // We'll stream the response
	})
	if err != nil {
		return "", fmt.Errorf("could not marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("could not create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned non-200 status: %s", resp.Status)
	}

	var summaryBuilder strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var lineResponse ollamaGenerateResponse
		if err := json.Unmarshal(scanner.Bytes(), &lineResponse); err != nil {
			// Ignore lines that are not valid JSON
			continue
		}
		summaryBuilder.WriteString(lineResponse.Response)
		if lineResponse.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading ollama stream: %w", err)
	}

	return strings.TrimSpace(summaryBuilder.String()), nil
}
