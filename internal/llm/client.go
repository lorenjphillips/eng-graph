package llm

import "context"

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type CompletionOptions struct {
	Temperature float64
	MaxTokens   int
	JSONMode    bool
}

type Client interface {
	Complete(ctx context.Context, messages []Message, opts CompletionOptions) (string, error)
	Stream(ctx context.Context, messages []Message, opts CompletionOptions) (<-chan string, <-chan error)
}
