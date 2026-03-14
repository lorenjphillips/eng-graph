package github

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/eng-graph/eng-graph/internal/source"
)

func init() {
	source.Register("github", func(name string, cfg map[string]any) (source.SourceAdapter, error) {
		a := &Adapter{name: name}

		if v, ok := cfg["token_env"].(string); ok {
			a.tokenEnv = v
		}
		if v, ok := cfg["user"].(string); ok {
			a.user = v
		}
		if v, ok := cfg["repos"].([]any); ok {
			for _, r := range v {
				if s, ok := r.(string); ok {
					a.repos = append(a.repos, s)
				}
			}
		}
		return a, nil
	})
}

type Adapter struct {
	name     string
	tokenEnv string
	user     string
	repos    []string
	client   *Client
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "github" }

func (a *Adapter) Validate() error {
	if a.tokenEnv == "" {
		return fmt.Errorf("github: token_env is required")
	}
	if len(a.repos) == 0 {
		return fmt.Errorf("github: repos is required")
	}
	for _, r := range a.repos {
		if !strings.Contains(r, "/") {
			return fmt.Errorf("github: repo %q must be in owner/repo format", r)
		}
	}
	if a.user == "" {
		return fmt.Errorf("github: user is required")
	}
	return nil
}

func (a *Adapter) TestConnection(ctx context.Context) error {
	token := os.Getenv(a.tokenEnv)
	if token == "" {
		return fmt.Errorf("github: env var %s is not set", a.tokenEnv)
	}
	a.client = NewClient(token)
	_, err := a.client.GetAuthenticatedUser(ctx)
	if err != nil {
		return fmt.Errorf("github: authentication failed: %w", err)
	}
	return nil
}

func (a *Adapter) Ingest(ctx context.Context, opts source.IngestOptions, out chan<- source.DataPoint, progress chan<- source.IngestProgress) error {
	if a.client == nil {
		token := os.Getenv(a.tokenEnv)
		if token == "" {
			return fmt.Errorf("github: env var %s is not set", a.tokenEnv)
		}
		a.client = NewClient(token)
	}

	total := 0
	for _, repo := range a.repos {
		parts := strings.SplitN(repo, "/", 2)
		owner, repoName := parts[0], parts[1]

		progress <- source.IngestProgress{
			Source:  a.name,
			Fetched: total,
			Message: fmt.Sprintf("fetching PR reviews from %s", repo),
		}

		reviews, err := a.client.ListPRReviews(ctx, owner, repoName, a.user, opts.Since)
		if err != nil {
			return fmt.Errorf("github: listing reviews for %s: %w", repo, err)
		}
		for _, r := range reviews {
			out <- TransformPRReview(r, a.name)
			total++
		}

		progress <- source.IngestProgress{
			Source:  a.name,
			Fetched: total,
			Message: fmt.Sprintf("fetching PR review comments from %s", repo),
		}

		comments, err := a.client.ListPRReviewComments(ctx, owner, repoName, a.user, opts.Since)
		if err != nil {
			return fmt.Errorf("github: listing review comments for %s: %w", repo, err)
		}
		for _, c := range comments {
			out <- TransformPRReviewComment(c, a.name)
			total++
		}

		progress <- source.IngestProgress{
			Source:  a.name,
			Fetched: total,
			Message: fmt.Sprintf("fetching PRs from %s", repo),
		}

		prs, err := a.client.ListPRs(ctx, owner, repoName, a.user, opts.Since)
		if err != nil {
			return fmt.Errorf("github: listing PRs for %s: %w", repo, err)
		}
		for _, pr := range prs {
			out <- TransformPR(pr, a.name)
			total++
		}
	}

	progress <- source.IngestProgress{
		Source:  a.name,
		Fetched: total,
		Message: "done",
	}
	return nil
}
