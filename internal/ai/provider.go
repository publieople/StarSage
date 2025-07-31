package ai

import "context"

// Provider is the interface that all AI providers must implement.
type Provider interface {
	// Generate takes a prompt and returns a text-based response from the AI model.
	Generate(ctx context.Context, prompt string) (string, error)
}
