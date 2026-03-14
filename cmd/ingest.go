package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/ingest"
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

var ingestCmd = &cobra.Command{
	Use:   "ingest <profile>",
	Short: "Ingest data from configured sources into a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runIngest,
}

var (
	ingestSource string
	ingestSince  string
	ingestAuthor string
)

func init() {
	ingestCmd.Flags().StringVar(&ingestSource, "source", "", "filter by source name")
	ingestCmd.Flags().StringVar(&ingestSince, "since", "", "only ingest data after this date (2024-01-01)")
	ingestCmd.Flags().StringVar(&ingestAuthor, "author", "", "override author filter")
	rootCmd.AddCommand(ingestCmd)
}

func runIngest(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	cfg, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	var sources []config.SourceConfig
	if ingestSource != "" {
		sc := cfg.FindSource(ingestSource)
		if sc == nil {
			return fmt.Errorf("source %q not found", ingestSource)
		}
		sources = []config.SourceConfig{*sc}
	} else {
		sources = cfg.Sources
	}

	if len(sources) == 0 {
		return fmt.Errorf("no sources configured")
	}

	var adapters []source.SourceAdapter
	for _, sc := range sources {
		a, err := source.Create(sc.Kind, sc.Name, sc.Config)
		if err != nil {
			return fmt.Errorf("creating adapter %q: %w", sc.Name, err)
		}
		if err := a.Validate(); err != nil {
			return fmt.Errorf("validating adapter %q: %w", sc.Name, err)
		}
		adapters = append(adapters, a)
	}

	opts := source.IngestOptions{Author: ingestAuthor}
	if ingestSince != "" {
		t, err := time.Parse("2006-01-02", ingestSince)
		if err != nil {
			return fmt.Errorf("parsing --since: %w", err)
		}
		opts.Since = t
	}

	dataDir := filepath.Join(dir, config.DirName, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, profileName+".db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pipeline := ingest.NewPipeline(store)
	total, err := pipeline.Run(ctx, profileName, adapters, opts)
	if err != nil {
		return fmt.Errorf("ingest failed: %w", err)
	}

	fmt.Printf("Ingested %d data points into profile %q\n", total, profileName)
	return nil
}
