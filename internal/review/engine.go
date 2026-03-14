package review

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
)

type Engine struct {
	client llm.Client
}

func NewEngine(client llm.Client) *Engine {
	return &Engine{client: client}
}

func (e *Engine) Review(ctx context.Context, p *profile.Profile, prRef string, format string) (string, error) {
	owner, repo, number, err := ParsePRRef(prRef)
	if err != nil {
		return "", err
	}

	if owner == "" || repo == "" {
		return "", fmt.Errorf("bare PR number requires owner/repo context: use owner/repo#%d or full URL", number)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	diff, err := FetchDiff(ctx, token, owner, repo, number)
	if err != nil {
		return "", fmt.Errorf("fetching diff: %w", err)
	}

	messages := BuildReviewPrompt(p, diff)

	if format == "json" {
		messages = append(messages, llm.Message{
			Role:    llm.RoleUser,
			Content: "Respond with a JSON object containing: summary (string), comments (array of {path, line, severity, body}), verdict (approve/request_changes/comment).",
		})
	} else if format == "gh" {
		messages = append(messages, llm.Message{
			Role:    llm.RoleUser,
			Content: "Format your response as a GitHub PR review. Start with a summary, then list inline comments with the format:\n\n**`file/path.go` L42**: comment text\n\nEnd with your verdict: APPROVE, REQUEST_CHANGES, or COMMENT.",
		})
	}

	opts := llm.CompletionOptions{
		Temperature: 0.3,
		MaxTokens:   4096,
		JSONMode:    format == "json",
	}

	result, err := e.client.Complete(ctx, messages, opts)
	if err != nil {
		return "", fmt.Errorf("LLM completion: %w", err)
	}

	if format == "json" {
		var check json.RawMessage
		if err := json.Unmarshal([]byte(result), &check); err != nil {
			return "", fmt.Errorf("LLM returned invalid JSON: %w", err)
		}
	}

	return result, nil
}
