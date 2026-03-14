package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/spf13/cobra"
)

var initAuto bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new eng-graph project",
	Long:  "Create .eng-graph/ directory. With --auto, detects gh CLI and git remote to auto-configure a GitHub source.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initAuto, "auto", false, "auto-detect GitHub CLI and git remote")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if err := config.Init(dir); err != nil {
		return fmt.Errorf("initializing project: %w", err)
	}
	fmt.Printf("Initialized eng-graph project in %s/%s\n", dir, config.DirName)

	if !initAuto {
		return nil
	}

	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}

	// Auto-detect GitHub via gh CLI.
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Println("gh CLI not found, skipping GitHub auto-detect")
		return nil
	}

	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		fmt.Println("gh not authenticated, skipping GitHub auto-detect")
		return nil
	}

	// Detect current repo from git remote.
	var repo string
	if out, err := exec.Command("git", "remote", "get-url", "origin").Output(); err == nil {
		repo = parseRepoFromRemote(strings.TrimSpace(string(out)))
	}

	// Detect username from gh.
	var user string
	if out, err := exec.Command("gh", "api", "/user", "--jq", ".login").Output(); err == nil {
		user = strings.TrimSpace(string(out))
	}

	if repo == "" {
		fmt.Println("Could not detect git remote, skipping GitHub auto-config")
		return nil
	}

	sc := config.SourceConfig{
		Name: "github",
		Kind: "github",
		Config: map[string]any{
			"token_env": "GITHUB_TOKEN",
			"repos":     []any{repo},
		},
	}
	if user != "" {
		sc.Config["user"] = user
	}

	cfg.Sources = append(cfg.Sources, sc)
	if err := config.Save(cfg, dir); err != nil {
		return err
	}

	msg := fmt.Sprintf("Auto-configured GitHub source: repo=%s", repo)
	if user != "" {
		msg += fmt.Sprintf(", user=%s", user)
	}
	fmt.Println(msg)
	return nil
}

func parseRepoFromRemote(remote string) string {
	// Handle SSH: git@github.com:org/repo.git
	if strings.HasPrefix(remote, "git@github.com:") {
		r := strings.TrimPrefix(remote, "git@github.com:")
		r = strings.TrimSuffix(r, ".git")
		return r
	}
	// Handle HTTPS: https://github.com/org/repo.git
	if strings.Contains(remote, "github.com/") {
		idx := strings.Index(remote, "github.com/")
		r := remote[idx+len("github.com/"):]
		r = strings.TrimSuffix(r, ".git")
		return r
	}
	return ""
}
