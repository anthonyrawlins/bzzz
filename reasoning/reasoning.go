package reasoning

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	ollamaAPIURL    = "http://localhost:11434/api/generate"
	defaultTimeout  = 60 * time.Second
)

// OllamaRequest represents the request payload for the Ollama API.
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a single streamed response object from the Ollama API.
type OllamaResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

// GenerateResponse queries the Ollama API with a given prompt and model,
// and returns the complete generated response as a single string.
func GenerateResponse(ctx context.Context, model, prompt string) (string, error) {
	// Set up a timeout for the request
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Create the request payload
	requestPayload := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false, // We will handle the full response at once for simplicity
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", ollamaAPIURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute http request to ollama: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama api returned non-200 status: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the JSON response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}
