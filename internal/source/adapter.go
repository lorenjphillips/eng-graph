package source

import (
	"context"
	"time"
)

type DataPointKind string

const (
	KindPRReviewComment DataPointKind = "pr_review_comment"
	KindPRReview        DataPointKind = "pr_review"
	KindPRDescription   DataPointKind = "pr_description"
	KindDocComment      DataPointKind = "doc_comment"
	KindMessage         DataPointKind = "message"
	KindNote            DataPointKind = "note"
)

type DataPoint struct {
	ID        string            `json:"id"`
	Source    string            `json:"source"`
	Kind      DataPointKind     `json:"kind"`
	Author    string            `json:"author"`
	Body      string            `json:"body"`
	Timestamp time.Time         `json:"timestamp"`
	Context   map[string]string `json:"context,omitempty"`
}

type IngestOptions struct {
	Author string
	Since  time.Time
}

type IngestProgress struct {
	Source  string
	Fetched int
	Message string
}

type SourceAdapter interface {
	Name() string
	Kind() string
	TestConnection(ctx context.Context) error
	Ingest(ctx context.Context, opts IngestOptions, out chan<- DataPoint, progress chan<- IngestProgress) error
	Validate() error
}
