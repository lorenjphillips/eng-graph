package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/spf13/cobra"
)

var sourceAddCmd = &cobra.Command{
	Use:   "add <kind>",
	Short: "Add a new data source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kind := args[0]

		cfg, err := config.Load(dir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Adding %s source\n", kind)
		fmt.Print("Source name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			name = kind
		}

		if cfg.FindSource(name) != nil {
			return fmt.Errorf("source %q already exists", name)
		}

		sc := config.SourceConfig{
			Name:   name,
			Kind:   kind,
			Config: make(map[string]any),
		}

		switch kind {
		case "github":
			fmt.Print("GitHub token env var [GITHUB_TOKEN]: ")
			tokenEnv, _ := reader.ReadString('\n')
			tokenEnv = strings.TrimSpace(tokenEnv)
			if tokenEnv == "" {
				tokenEnv = "GITHUB_TOKEN"
			}
			sc.Config["token_env"] = tokenEnv

			fmt.Print("Repos (comma-separated, e.g. owner/repo1,owner/repo2): ")
			reposLine, _ := reader.ReadString('\n')
			reposLine = strings.TrimSpace(reposLine)
			var repos []any
			for _, r := range strings.Split(reposLine, ",") {
				r = strings.TrimSpace(r)
				if r != "" {
					repos = append(repos, r)
				}
			}
			sc.Config["repos"] = repos

			fmt.Print("GitHub username: ")
			user, _ := reader.ReadString('\n')
			sc.Config["user"] = strings.TrimSpace(user)

		default:
			fmt.Println("Enter key=value pairs (empty line to finish):")
			for {
				line, _ := reader.ReadString('\n')
				line = strings.TrimSpace(line)
				if line == "" {
					break
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					sc.Config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		cfg.Sources = append(cfg.Sources, sc)
		if err := config.Save(cfg, dir); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Source %q (%s) added\n", name, kind)
		return nil
	},
}

func init() {
	sourceCmd.AddCommand(sourceAddCmd)
}
