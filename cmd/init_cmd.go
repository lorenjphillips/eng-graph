package cmd

import (
	"fmt"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new eng-graph project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Init(dir); err != nil {
			return fmt.Errorf("initializing project: %w", err)
		}
		fmt.Printf("Initialized eng-graph project in %s/%s\n", dir, config.DirName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
