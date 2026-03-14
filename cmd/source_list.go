package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/spf13/cobra"
)

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured data sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(dir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if len(cfg.Sources) == 0 {
			fmt.Println("No sources configured. Run 'eng-graph source add <kind>' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tKIND\tSUMMARY")
		for _, s := range cfg.Sources {
			fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Kind, summarizeConfig(s.Config))
		}
		return w.Flush()
	},
}

func summarizeConfig(cfg map[string]any) string {
	var parts []string
	for k, v := range cfg {
		switch val := v.(type) {
		case []any:
			parts = append(parts, fmt.Sprintf("%s: %d items", k, len(val)))
		default:
			parts = append(parts, fmt.Sprintf("%s: %v", k, val))
		}
	}
	return strings.Join(parts, ", ")
}

func init() {
	sourceCmd.AddCommand(sourceListCmd)
}
