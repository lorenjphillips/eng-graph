package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/interview"
	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/spf13/cobra"
)

var interviewCmd = &cobra.Command{
	Use:   "interview <profile>",
	Short: "Run an interactive interview to enrich a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runInterview,
}

var (
	interviewResume     bool
	interviewTranscript string
)

func init() {
	interviewCmd.Flags().BoolVar(&interviewResume, "resume", false, "resume a previous interview")
	interviewCmd.Flags().StringVar(&interviewTranscript, "transcript", "", "import transcript from file")
	rootCmd.AddCommand(interviewCmd)
}

func runInterview(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	cfg, err := config.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	store := profile.NewStore(dir)
	p, err := store.Load(profileName)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	transcriptPath := filepath.Join(store.ProfileDir(profileName), "interview.json")

	if interviewTranscript != "" {
		data, err := os.ReadFile(interviewTranscript)
		if err != nil {
			return fmt.Errorf("reading transcript: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(transcriptPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(transcriptPath, data, 0644); err != nil {
			return err
		}
		fmt.Printf("Imported transcript to %s\n", transcriptPath)
		return nil
	}

	client, err := llm.NewOpenAIClient(cfg.LLM)
	if err != nil {
		return fmt.Errorf("creating LLM client: %w", err)
	}

	conductor := interview.NewConductor(client, p)

	if interviewResume {
		if t, err := interview.LoadTranscript(transcriptPath); err == nil {
			conductor.LoadHistory(t.Exchanges)
			fmt.Fprintf(os.Stderr, "Resuming interview (%d questions answered)\n", len(t.Exchanges))
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	transcript, err := interview.RunInterview(ctx, conductor)
	if err != nil {
		return fmt.Errorf("interview: %w", err)
	}

	if err := transcript.Save(transcriptPath); err != nil {
		return fmt.Errorf("saving transcript: %w", err)
	}

	fmt.Printf("Interview saved (%d exchanges) to %s\n", len(transcript.Exchanges), transcriptPath)
	return nil
}
