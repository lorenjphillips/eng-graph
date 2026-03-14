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
	Long: `Add a new data source. For non-interactive use, pass all config via flags:

  eng-graph source add github --name my-gh --token-env GITHUB_TOKEN --repos org/repo1,org/repo2 --user myuser
  eng-graph source add gitlab --name my-gl --config token_env=GITLAB_TOKEN --config project=mygroup/myproject`,
	Args: cobra.ExactArgs(1),
	RunE: runSourceAdd,
}

var (
	sourceAddName     string
	sourceAddTokenEnv string
	sourceAddRepos    string
	sourceAddUser     string
	sourceAddConfigs  []string
)

func init() {
	sourceAddCmd.Flags().StringVar(&sourceAddName, "name", "", "source name (default: auto-generated from kind)")
	sourceAddCmd.Flags().StringVar(&sourceAddTokenEnv, "token-env", "", "env var name for API token")
	sourceAddCmd.Flags().StringVar(&sourceAddRepos, "repos", "", "comma-separated repos (owner/repo)")
	sourceAddCmd.Flags().StringVar(&sourceAddUser, "user", "", "username to filter by")
	sourceAddCmd.Flags().StringArrayVar(&sourceAddConfigs, "config", nil, "key=value config pairs")
	sourceCmd.AddCommand(sourceAddCmd)
}

func runSourceAdd(cmd *cobra.Command, args []string) error {
	kind := args[0]

	cfg, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	nonInteractive := sourceAddTokenEnv != "" || sourceAddRepos != "" || sourceAddUser != "" || len(sourceAddConfigs) > 0

	if nonInteractive {
		return addSourceFromFlags(cfg, kind)
	}
	return addSourceInteractive(cfg, kind)
}

func addSourceFromFlags(cfg *config.Config, kind string) error {
	name := sourceAddName
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

	if sourceAddTokenEnv != "" {
		sc.Config["token_env"] = sourceAddTokenEnv
	}
	if sourceAddUser != "" {
		sc.Config["user"] = sourceAddUser
	}
	if sourceAddRepos != "" {
		var repos []any
		for _, r := range strings.Split(sourceAddRepos, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				repos = append(repos, r)
			}
		}
		sc.Config["repos"] = repos
	}
	for _, kv := range sourceAddConfigs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			sc.Config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	cfg.Sources = append(cfg.Sources, sc)
	if err := config.Save(cfg, dir); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Source %q (%s) added\n", name, kind)
	return nil
}

func addSourceInteractive(cfg *config.Config, kind string) error {
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
}
