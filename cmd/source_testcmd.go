package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/source"
	"github.com/spf13/cobra"

	_ "github.com/eng-graph/eng-graph/internal/source/confluence"
	_ "github.com/eng-graph/eng-graph/internal/source/gdocs"
	_ "github.com/eng-graph/eng-graph/internal/source/github"
	_ "github.com/eng-graph/eng-graph/internal/source/gitlab"
	_ "github.com/eng-graph/eng-graph/internal/source/notion"
	_ "github.com/eng-graph/eng-graph/internal/source/obsidian"
	_ "github.com/eng-graph/eng-graph/internal/source/quip"
)

var sourceTestCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test connectivity for a configured source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load(dir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		sc := cfg.FindSource(name)
		if sc == nil {
			return fmt.Errorf("source %q not found", name)
		}

		adapter, err := source.Create(sc.Kind, sc.Name, sc.Config)
		if err != nil {
			return fmt.Errorf("creating adapter: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fmt.Printf("Testing connection to %s (%s)...\n", sc.Name, sc.Kind)
		if err := adapter.TestConnection(ctx); err != nil {
			fmt.Printf("FAIL: %v\n", err)
			return err
		}

		fmt.Println("OK")
		return nil
	},
}

func init() {
	sourceCmd.AddCommand(sourceTestCmd)
}
