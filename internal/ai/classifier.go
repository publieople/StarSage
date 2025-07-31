package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"star-sage/internal/db"
	"strings"
)

const (
	// maxTokensPerChunk is an approximate limit to avoid exceeding model context windows.
	// This should be adjusted based on the specific model's limits and average token usage per item.
	maxTokensPerChunk = 4096
)

// ClassifyRepositories uses an AI provider to classify repositories based on a user prompt.
// It handles chunking the repositories to fit within the AI model's context window.
func ClassifyRepositories(ctx context.Context, provider Provider, userPrompt string, repos []db.Repository) ([]int64, error) {
	var finalRepoIDs []int64

	chunks, err := chunkRepositories(repos)
	if err != nil {
		return nil, fmt.Errorf("could not chunk repositories: %w", err)
	}

	for i, chunk := range chunks {
		fmt.Printf("Processing chunk %d/%d...\n", i+1, len(chunks))
		prompt, err := buildClassificationPrompt(userPrompt, chunk)
		if err != nil {
			return nil, fmt.Errorf("could not build prompt for chunk %d: %w", i, err)
		}

		resp, err := provider.Generate(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("ai generation failed for chunk %d: %w", i, err)
		}

		repoIDs, err := parseAIResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("could not parse AI response for chunk %d: %w", i, err)
		}
		finalRepoIDs = append(finalRepoIDs, repoIDs...)
	}

	return finalRepoIDs, nil
}

// chunkRepositories splits a slice of repositories into smaller chunks based on estimated token count.
func chunkRepositories(repos []db.Repository) ([][]db.Repository, error) {
	var chunks [][]db.Repository
	var currentChunk []db.Repository
	var currentTokens int

	for _, repo := range repos {
		// Simple token estimation: 1 token ~ 4 chars. This is a rough heuristic.
		repoTokens := (len(repo.FullName) + len(repo.Description) + len(repo.Summary)) / 4

		if currentTokens+repoTokens > maxTokensPerChunk && len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
			currentChunk = nil
			currentTokens = 0
		}

		currentChunk = append(currentChunk, repo)
		currentTokens += repoTokens
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks, nil
}

// buildClassificationPrompt creates the full prompt to be sent to the AI model.
func buildClassificationPrompt(userPrompt string, repos []db.Repository) (string, error) {
	type repoInfo struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Summary     string `json:"summary,omitempty"`
	}

	var infos []repoInfo
	for _, r := range repos {
		infos = append(infos, repoInfo{
			ID:          r.ID,
			Name:        r.FullName,
			Description: r.Description,
			Summary:     r.Summary,
		})
	}

	jsonData, err := json.Marshal(infos)
	if err != nil {
		return "", fmt.Errorf("could not marshal repo info to JSON: %w", err)
	}

	promptTemplate := `你是一个精准的软件项目分类助手。
我会给你一个分类任务的描述，以及一个 JSON 格式的项目列表。
请仔细阅读每个项目的描述，并判断它是否符合分类任务的要求。

分类任务: "%s"

项目列表如下:
%s

请只返回一个 JSON 数组，其中仅包含符合分类任务要求的项目 ID。
例如: [12345, 67890]
确保你的回答中除了这个 JSON 数组外，不包含任何其他文字、解释或代码块标记。`

	return fmt.Sprintf(promptTemplate, userPrompt, string(jsonData)), nil
}

// parseAIResponse extracts the JSON array of repository IDs from the AI's text response.
func parseAIResponse(response string) ([]int64, error) {
	// The AI might sometimes wrap the JSON in markdown code blocks.
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var repoIDs []int64
	err := json.Unmarshal([]byte(response), &repoIDs)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal AI response JSON: %w (response was: %s)", err, response)
	}

	return repoIDs, nil
}
