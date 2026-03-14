package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/eng-graph/eng-graph/internal/review"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review <pr-ref>",
	Short: "Review a PR using an engineer profile",
	Long:  "Accepts a GitHub PR URL, owner/repo#number, or PR number (requires config for repo).",
	Args:  cobra.ExactArgs(1),
	RunE:  runReview,
}

var (
	reviewProfile string
	reviewFormat  string
)

func init() {
	reviewCmd.Flags().StringVar(&reviewProfile, "profile", "", "profile name (default: active_profile from config)")
	reviewCmd.Flags().StringVar(&reviewFormat, "format", "text", "output format (text, json, gh)")
	rootCmd.AddCommand(reviewCmd)
}

func runReview(cmd *cobra.Command, args []string) error {
	prRef := args[0]

	cfg, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	profileName := reviewProfile
	if profileName == "" {
		profileName = cfg.ActiveProfile
	}
	if profileName == "" {
		return fmt.Errorf("no profile specified; use --profile or set active_profile in config")
	}

	store := profile.NewStore(dir)
	p, err := store.Load(profileName)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	owner, repo, number, err := review.ParsePRRef(prRef)
	if err != nil {
		return err
	}

	// For bare PR numbers, resolve owner/repo from first github source.
	if owner == "" {
		ghSources := cfg.SourceByKind("github")
		if len(ghSources) == 0 {
			return fmt.Errorf("bare PR number requires a github source in config")
		}
		repos, _ := ghSources[0].Config["repos"].([]any)
		if len(repos) == 0 {
			return fmt.Errorf("github source has no repos configured")
		}
		repoStr, _ := repos[0].(string)
		parts := strings.SplitN(repoStr, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("repo %q must be in owner/repo format", repoStr)
		}
		owner, repo = parts[0], parts[1]
	}

	// Resolve GitHub token from first github source.
	tokenEnv := "GITHUB_TOKEN"
	if ghSources := cfg.SourceByKind("github"); len(ghSources) > 0 {
		if env, ok := ghSources[0].Config["token_env"].(string); ok {
			tokenEnv = env
		}
	}
	token := os.Getenv(tokenEnv)
	if token == "" {
		return fmt.Errorf("environment variable %s is not set", tokenEnv)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	diff, err := review.FetchDiff(ctx, token, owner, repo, number)
	if err != nil {
		return fmt.Errorf("fetching PR diff: %w", err)
	}

	messages := review.BuildReviewPrompt(p, diff)

	client, err := llm.NewOpenAIClient(cfg.LLM)
	if err != nil {
		return fmt.Errorf("creating LLM client: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Reviewing %s (%d files)...\n", diff.URL, len(diff.Files))

	switch reviewFormat {
	case "text", "gh":
		chunks, errs := client.Stream(ctx, messages, llm.CompletionOptions{
			Temperature: 0.3,
			MaxTokens:   4096,
		})
		for chunk := range chunks {
			fmt.Print(chunk)
		}
		fmt.Println()
		if err := <-errs; err != nil {
			return fmt.Errorf("LLM stream: %w", err)
		}

	case "json":
		resp, err := client.Complete(ctx, messages, llm.CompletionOptions{
			Temperature: 0.3,
			MaxTokens:   4096,
		})
		if err != nil {
			return fmt.Errorf("LLM completion: %w", err)
		}
		out := map[string]any{
			"pr":     diff.URL,
			"review": resp,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)

	default:
		return fmt.Errorf("unknown format %q (supported: text, json, gh)", reviewFormat)
	}

	return nil
}
