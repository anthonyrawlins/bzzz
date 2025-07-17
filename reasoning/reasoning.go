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

var (
	availableModels []string
	modelWebhookURL string
	defaultModel    string
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

// SetModelConfig configures the available models and webhook URL for smart model selection
func SetModelConfig(models []string, webhookURL, defaultReasoningModel string) {
	availableModels = models
	modelWebhookURL = webhookURL
	defaultModel = defaultReasoningModel
}

// selectBestModel calls the model selection webhook to choose the best model for a prompt
func selectBestModel(availableModels []string, prompt string) string {
	if modelWebhookURL == "" || len(availableModels) == 0 {
		// Fallback to first available model
		if len(availableModels) > 0 {
			return availableModels[0]
		}
		return defaultModel // Last resort fallback
	}
	
	requestPayload := map[string]interface{}{
		"models": availableModels,
		"prompt": prompt,
	}
	
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		// Fallback on error
		return availableModels[0]
	}
	
	resp, err := http.Post(modelWebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		// Fallback on error
		return availableModels[0]
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		// Fallback on error
		return availableModels[0]
	}
	
	var response struct {
		Model string `json:"model"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		// Fallback on error
		return availableModels[0]
	}
	
	// Validate that the returned model is in our available list
	for _, model := range availableModels {
		if model == response.Model {
			return response.Model
		}
	}
	
	// Fallback if webhook returned invalid model
	return availableModels[0]
}

// GenerateResponseSmart automatically selects the best model for the prompt
func GenerateResponseSmart(ctx context.Context, prompt string) (string, error) {
	selectedModel := selectBestModel(availableModels, prompt)
	return GenerateResponse(ctx, selectedModel, prompt)
}
