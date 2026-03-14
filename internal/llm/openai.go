package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eng-graph/eng-graph/internal/config"
)

type OpenAIClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAIClient(cfg config.LLMConfig) (*OpenAIClient, error) {
	apiKey := os.Getenv(cfg.APIKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("environment variable %s is not set", cfg.APIKeyEnv)
	}
	return &OpenAIClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  apiKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

type chatRequest struct {
	Model          string            `json:"model"`
	Messages       []Message         `json:"messages"`
	Temperature    float64           `json:"temperature,omitempty"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	Stream         bool              `json:"stream,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type streamDelta struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (c *OpenAIClient) Complete(ctx context.Context, messages []Message, opts CompletionOptions) (string, error) {
	body := c.buildRequest(messages, opts, false)
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result chatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) Stream(ctx context.Context, messages []Message, opts CompletionOptions) (<-chan string, <-chan error) {
	chunks := make(chan string, 64)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		body := c.buildRequest(messages, opts, true)
		data, err := json.Marshal(body)
		if err != nil {
			errs <- fmt.Errorf("marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(data))
		if err != nil {
			errs <- fmt.Errorf("create request: %w", err)
			return
		}
		c.setHeaders(req)

		// Use a separate client without timeout for streaming.
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			errs <- fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				return
			}

			var delta streamDelta
			if err := json.Unmarshal([]byte(payload), &delta); err != nil {
				continue
			}
			if len(delta.Choices) > 0 && delta.Choices[0].Delta.Content != "" {
				select {
				case chunks <- delta.Choices[0].Delta.Content:
				case <-ctx.Done():
					errs <- ctx.Err()
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("read stream: %w", err)
		}
	}()

	return chunks, errs
}

func (c *OpenAIClient) buildRequest(messages []Message, opts CompletionOptions, stream bool) chatRequest {
	req := chatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stream:      stream,
	}
	if opts.JSONMode {
		req.ResponseFormat = &responseFormat{Type: "json_object"}
	}
	return req
}

func (c *OpenAIClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}
