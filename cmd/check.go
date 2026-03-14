package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var checkJSON bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check available CLIs and authentication status",
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&checkJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(checkCmd)
}

type checkResult struct {
	Tool      string `json:"tool"`
	Available bool   `json:"available"`
	Auth      string `json:"auth,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

func runCheck(cmd *cobra.Command, args []string) error {
	results := []checkResult{
		checkGH(),
		checkGit(),
		checkTool("notion", "notion"),
		checkTool("confluence", "confluence"),
	}

	if checkJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	for _, r := range results {
		status := "not found"
		if r.Available {
			status = "ok"
		}
		line := fmt.Sprintf("%-15s %s", r.Tool, status)
		if r.Auth != "" {
			line += fmt.Sprintf("  (auth: %s)", r.Auth)
		}
		if r.Detail != "" {
			line += fmt.Sprintf("  %s", r.Detail)
		}
		fmt.Println(line)
	}
	return nil
}

func checkGH() checkResult {
	r := checkResult{Tool: "gh"}
	if _, err := exec.LookPath("gh"); err != nil {
		return r
	}
	r.Available = true

	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		r.Auth = "not authenticated"
		return r
	}
	r.Auth = "authenticated"

	out, err = exec.Command("gh", "api", "/user", "--jq", ".login").Output()
	if err == nil {
		r.Detail = "user=" + strings.TrimSpace(string(out))
	}
	return r
}

func checkGit() checkResult {
	r := checkResult{Tool: "git"}
	if _, err := exec.LookPath("git"); err != nil {
		return r
	}
	r.Available = true

	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err == nil {
		r.Detail = "origin=" + strings.TrimSpace(string(out))
	}
	return r
}

func checkTool(name, bin string) checkResult {
	r := checkResult{Tool: name}
	if _, err := exec.LookPath(bin); err == nil {
		r.Available = true
	}
	return r
}
