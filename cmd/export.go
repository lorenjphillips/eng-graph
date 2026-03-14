package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <profile>",
	Short: "Export profile output to a directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

var (
	exportFormat string
	exportOutput string
)

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "md", "output format (md, json)")
	exportCmd.Flags().StringVar(&exportOutput, "output", ".", "target directory")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	store := profile.NewStore(dir)

	switch exportFormat {
	case "json":
		p, err := store.Load(profileName)
		if err != nil {
			return fmt.Errorf("loading profile: %w", err)
		}
		data, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		dst := filepath.Join(exportOutput, "profile.json")
		if err := os.MkdirAll(exportOutput, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return err
		}
		fmt.Printf("Exported profile JSON to %s\n", dst)

	case "md":
		srcDir := store.OutputDir(profileName)
		if err := os.MkdirAll(exportOutput, 0755); err != nil {
			return err
		}
		copied, err := copyDir(srcDir, exportOutput)
		if err != nil {
			return fmt.Errorf("copying output: %w", err)
		}
		fmt.Printf("Exported %d files to %s\n", copied, exportOutput)

	default:
		return fmt.Errorf("unknown format %q (supported: md, json)", exportFormat)
	}

	return nil
}

func copyDir(src, dst string) (int, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())

		if e.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return count, err
			}
			n, err := copyDir(srcPath, dstPath)
			if err != nil {
				return count, err
			}
			count += n
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}
