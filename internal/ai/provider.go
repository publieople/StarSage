package ai

import "context"

// Provider is the interface that all AI providers must implement.
type Provider interface {
	// Summarize takes a string of content (e.g., a README) and returns a summary.
	Summarize(ctx context.Context, content string) (string, error)
}
