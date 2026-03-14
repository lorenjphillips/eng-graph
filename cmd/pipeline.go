package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/ingest"
	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/eng-graph/eng-graph/internal/source"
	"github.com/eng-graph/eng-graph/internal/storage"
	"github.com/spf13/cobra"

	_ "github.com/eng-graph/eng-graph/internal/source/confluence"
	_ "github.com/eng-graph/eng-graph/internal/source/gdocs"
	_ "github.com/eng-graph/eng-graph/internal/source/github"
	_ "github.com/eng-graph/eng-graph/internal/source/gitlab"
	_ "github.com/eng-graph/eng-graph/internal/source/notion"
	_ "github.com/eng-graph/eng-graph/internal/source/obsidian"
	_ "github.com/eng-graph/eng-graph/internal/source/quip"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline <name>",
	Short: "Auto-detect, ingest, and dump a profile in one shot",
	Long: `One command to go from zero to data. Runs the full pipeline:

  1. Initialize .eng-graph/ if needed
  2. Auto-detect GitHub from gh CLI and git remote
  3. Create profile if it doesn't exist
  4. Ingest data from all configured sources
  5. Dump all ingested data as JSON to stdout

Designed for AI agents: pipe the output directly into your analysis.

  eng-graph pipeline emily --since 2025-01-01`,
	Args: cobra.ExactArgs(1),
	RunE: runPipeline,
}

var (
	pipelineSince string
	pipelineLimit int
	pipelineRepos []string
	pipelineUser  string
)

func init() {
	pipelineCmd.Flags().StringVar(&pipelineSince, "since", "2025-01-01", "ingest data since this date (YYYY-MM-DD)")
	pipelineCmd.Flags().IntVar(&pipelineLimit, "limit", 0, "max data points to dump (0 = all)")
	pipelineCmd.Flags().StringSliceVar(&pipelineRepos, "repos", nil, "GitHub repos to ingest (owner/repo)")
	pipelineCmd.Flags().StringVar(&pipelineUser, "user", "", "GitHub username to filter by")
	rootCmd.AddCommand(pipelineCmd)
}

func runPipeline(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Step 1: Init if needed.
	cfgPath := filepath.Join(dir, config.DirName, "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := config.Init(dir); err != nil {
			return fmt.Errorf("initializing: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Initialized .eng-graph/")
	}

	cfg, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Step 2: Auto-detect GitHub if no sources configured.
	if len(cfg.Sources) == 0 {
		sc, err := autoDetectGitHub()
		if err != nil {
			return fmt.Errorf("auto-detect: %w", err)
		}
		cfg.Sources = append(cfg.Sources, *sc)
		if err := config.Save(cfg, dir); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Auto-configured GitHub source: %v\n", sc.Config["repos"])
	}

	// Apply flag overrides.
	if len(pipelineRepos) > 0 || pipelineUser != "" {
		for i := range cfg.Sources {
			if cfg.Sources[i].Kind == "github" {
				if len(pipelineRepos) > 0 {
					repos := make([]any, len(pipelineRepos))
					for j, r := range pipelineRepos {
						repos[j] = r
					}
					cfg.Sources[i].Config["repos"] = repos
				}
				if pipelineUser != "" {
					cfg.Sources[i].Config["user"] = pipelineUser
				}
			}
		}
	}

	// Step 3: Create profile if needed.
	store := profile.NewStore(dir)
	if _, err := store.Load(profileName); err != nil {
		p := &profile.Profile{Name: profileName, DisplayName: profileName}
		if err := store.Save(p); err != nil {
			return fmt.Errorf("creating profile: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Created profile %q\n", profileName)
	}

	// Step 4: Ingest.
	since, err := time.Parse("2006-01-02", pipelineSince)
	if err != nil {
		return fmt.Errorf("invalid --since date: %w", err)
	}

	var adapters []source.SourceAdapter
	for _, sc := range cfg.Sources {
		a, err := source.Create(sc.Kind, sc.Name, sc.Config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Skipping source %s: %v\n", sc.Name, err)
			continue
		}
		if err := a.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Skipping source %s: %v\n", sc.Name, err)
			continue
		}
		adapters = append(adapters, a)
	}

	if len(adapters) == 0 {
		return fmt.Errorf("no valid sources configured")
	}

	dbPath := filepath.Join(dir, config.DirName, "data", profileName+".db")
	db, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	opts := source.IngestOptions{Since: since, Author: pipelineUser}
	pipeline := ingest.NewPipeline(db)
	total, err := pipeline.Run(ctx, profileName, adapters, opts)
	if err != nil {
		return fmt.Errorf("ingest: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Ingested %d data points\n", total)

	// Step 5: Dump.
	dps, err := db.LoadAll(profileName)
	if err != nil {
		return fmt.Errorf("loading data: %w", err)
	}

	if pipelineLimit > 0 && pipelineLimit < len(dps) {
		dps = dps[:pipelineLimit]
	}

	fmt.Fprintf(os.Stderr, "Dumping %d data points to stdout\n", len(dps))

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(dps)
}

func autoDetectGitHub() (*config.SourceConfig, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI not found; install it or configure sources manually")
	}

	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return nil, fmt.Errorf("gh CLI not authenticated; run 'gh auth login'")
	}

	var repo string
	if out, err := exec.Command("git", "remote", "get-url", "origin").Output(); err == nil {
		repo = parseRepoFromRemote(strings.TrimSpace(string(out)))
	}
	if repo == "" {
		return nil, fmt.Errorf("could not detect git remote; use --repos flag")
	}

	var user string
	if out, err := exec.Command("gh", "api", "/user", "--jq", ".login").Output(); err == nil {
		user = strings.TrimSpace(string(out))
	}

	sc := &config.SourceConfig{
		Name: "github",
		Kind: "github",
		Config: map[string]any{
			"token_env": "GITHUB_TOKEN",
			"repos":     []any{repo},
		},
	}
	if user != "" {
		sc.Config["user"] = user
	}
	return sc, nil
}
