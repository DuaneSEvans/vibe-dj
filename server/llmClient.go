package main

import (
	"context"
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
