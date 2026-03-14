package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/eng-graph/eng-graph/internal/builder"
	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/eng-graph/eng-graph/internal/storage"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build <profile>",
	Short: "Build a profile from ingested data using a local LLM",
	Long: `Analyze ingested data and synthesize a tiered profile using an LLM.

Requires LLM configuration in .eng-graph/config.yaml. Most users should
use 'eng-graph dump' instead and let their AI agent analyze the data directly.`,
	Args: cobra.ExactArgs(1),
	RunE: runBuild,
}

var (
	buildTier   int
	buildOutput string
)

func init() {
	buildCmd.Flags().IntVar(&buildTier, "tier", 0, "build specific tier (1-4, default: all)")
	buildCmd.Flags().StringVar(&buildOutput, "output", "", "output directory (default: profile output dir)")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	cfg, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.LLM.APIKeyEnv == "" {
		return fmt.Errorf("no LLM configured. Use 'eng-graph dump %s' to export data for your AI agent to analyze directly", profileName)
	}

	store := profile.NewStore(dir)
	p, err := store.Load(profileName)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	dbPath := filepath.Join(dir, config.DirName, "data", profileName+".db")
	db, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	datapoints, err := db.LoadAll(profileName)
	if err != nil {
		return fmt.Errorf("loading data points: %w", err)
	}
	if len(datapoints) == 0 {
		return fmt.Errorf("no data points found for profile %q; run 'eng-graph ingest %s' first", profileName, profileName)
	}

	var transcript json.RawMessage
	transcriptPath := filepath.Join(store.ProfileDir(profileName), "interview.json")
	if data, err := os.ReadFile(transcriptPath); err == nil {
		transcript = data
	}

	client, err := llm.NewOpenAIClient(cfg.LLM)
	if err != nil {
		return fmt.Errorf("creating LLM client: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b := builder.NewBuilder(client)
	if err := b.Build(ctx, p, datapoints, transcript); err != nil {
		return fmt.Errorf("building profile: %w", err)
	}

	if err := store.Save(p); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	outputDir := buildOutput
	if outputDir == "" {
		outputDir = store.OutputDir(profileName)
	}

	if err := builder.Render(p, outputDir); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	fmt.Printf("Profile %q built and rendered to %s\n", profileName, outputDir)
	return nil
}
