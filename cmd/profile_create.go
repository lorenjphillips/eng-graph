package cmd

import (
	"fmt"

	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/spf13/cobra"
)

var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new engineer profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store := profile.NewStore(dir)
		p, err := store.Create(args[0])
		if err != nil {
			return fmt.Errorf("creating profile: %w", err)
		}
		fmt.Printf("Created profile %q\n", p.Name)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileCreateCmd)
}
