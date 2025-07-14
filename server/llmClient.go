package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

/**
 * LLMClient defines the interface for interacting with a large language model.
 * In production, the server will interact with a hosted service to avoid the
 * costs of self hosting a machine that can handle the LLM. Locally, however, to
 * save costs, the server uses a local Ollama instance for development.
 */
type LLMClient interface {
	DescribeImage(ctx context.Context, imageData []byte, prompt string) (string, error)
}

// #region Ollama
type ollamaGenerateRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images"`
	Stream bool     `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

type OllamaClient struct {
	httpClient *http.Client
	hostURL    string
}

func NewOllamaClient(hostURL string) *OllamaClient {
	return &OllamaClient{
		httpClient: &http.Client{},
		hostURL:    hostURL,
	}
}

func (c *OllamaClient) DescribeImage(ctx context.Context, imageData []byte, prompt string) (string, error) {
	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	payload := ollamaGenerateRequest{
		Model:  "llava",
		Prompt: prompt,
		Images: []string{encodedImage},
		Stream: false,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.hostURL+"/api/generate", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned a non-200 status code: %d", resp.StatusCode)
	}

	var ollamaResp ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return ollamaResp.Response, nil
}

// #endregion

// #region Production, hosted
type ReplicateClient struct {
	apiToken string
}

func NewReplicateClient(apiToken string) *ReplicateClient {
	return &ReplicateClient{
		apiToken: apiToken,
	}
}

func (c *ReplicateClient) DescribeImage(ctx context.Context, imageData []byte, prompt string) (string, error) {
	// ... logic to maybe upload image to a temp store like S3
	// ... logic to build the Replicate-specific JSON request
	// ... make http POST to api.replicate.com/v1/predictions
	// ... poll the status url until the prediction is complete
	// ... parse the Replicate-specific response
	return "description from replicate", nil
}

// #endregion
