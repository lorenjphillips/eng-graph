package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/storage"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump <profile>",
	Short: "Dump raw ingested data as JSON to stdout",
	Long:  "Output all ingested data points for a profile as a JSON array. Designed for agents to read and analyze directly.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDump,
}

var dumpLimit int

func init() {
	dumpCmd.Flags().IntVar(&dumpLimit, "limit", 0, "max data points to output (0 = all)")
	rootCmd.AddCommand(dumpCmd)
}

func runDump(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	dbPath := filepath.Join(dir, config.DirName, "data", profileName+".db")
	db, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	dps, err := db.LoadAll(profileName)
	if err != nil {
		return fmt.Errorf("loading data points: %w", err)
	}

	if len(dps) == 0 {
		return fmt.Errorf("no data points for profile %q", profileName)
	}

	if dumpLimit > 0 && dumpLimit < len(dps) {
		dps = dps[:dumpLimit]
	}

	fmt.Fprintf(os.Stderr, "%d data points\n", len(dps))

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(dps)
}
