package cmd

import (
	"fmt"

	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := profile.NewStore(dir)
		names, err := store.List()
		if err != nil {
			return fmt.Errorf("listing profiles: %w", err)
		}
		if len(names) == 0 {
			fmt.Println("No profiles found. Run 'eng-graph profile create <name>' to create one.")
			return nil
		}
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileListCmd)
}
