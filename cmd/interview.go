package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/eng-graph/eng-graph/internal/config"
	"github.com/eng-graph/eng-graph/internal/interview"
	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/spf13/cobra"
)

var interviewCmd = &cobra.Command{
	Use:   "interview <profile>",
	Short: "Run an interactive interview to enrich a profile",
	Long: `Run an AI-driven interview to capture engineering preferences.

By default, opens a terminal UI. For non-interactive use (scripts, agents):

  eng-graph interview myprofile --batch
  eng-graph interview myprofile --transcript answers.json`,
	Args: cobra.ExactArgs(1),
	RunE: runInterview,
}

var (
	interviewResume     bool
	interviewTranscript string
	interviewBatch      bool
)

func init() {
	interviewCmd.Flags().BoolVar(&interviewResume, "resume", false, "resume a previous interview")
	interviewCmd.Flags().StringVar(&interviewTranscript, "transcript", "", "import transcript from file")
	interviewCmd.Flags().BoolVar(&interviewBatch, "batch", false, "non-interactive mode: print questions to stdout, read answers from stdin")
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

	var transcript *interview.Transcript

	if interviewBatch {
		transcript, err = runBatchInterview(ctx, conductor)
	} else {
		transcript, err = interview.RunInterview(ctx, conductor)
	}
	if err != nil {
		return fmt.Errorf("interview: %w", err)
	}

	if err := transcript.Save(transcriptPath); err != nil {
		return fmt.Errorf("saving transcript: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Interview saved (%d exchanges) to %s\n", len(transcript.Exchanges), transcriptPath)
	return nil
}

func runBatchInterview(ctx context.Context, conductor *interview.Conductor) (*interview.Transcript, error) {
	reader := bufio.NewReader(os.Stdin)

	for !conductor.IsComplete() {
		question, err := conductor.NextQuestion(ctx)
		if err != nil {
			return nil, err
		}
		if question == "" {
			break
		}

		fmt.Printf("QUESTION: %s\n", question)

		answer, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		answer = strings.TrimSpace(answer)
		if answer == "" || strings.EqualFold(answer, "quit") {
			break
		}

		conductor.SubmitAnswer(answer)
	}

	return conductor.Transcript(), nil
}
