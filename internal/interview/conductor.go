package interview

import (
	"context"
	"time"

	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
)

type Phase int

const (
	PhaseIntro Phase = iota
	PhasePriorities
	PhaseCodePatterns
	PhaseArchitecture
	PhaseErrors
	PhaseDatabase
	PhaseCommunication
	PhaseWrapUp
)

const (
	minQuestions      = 15
	maxQuestions      = 25
	questionsPerPhase = 3
)

type Exchange struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Phase    Phase  `json:"phase"`
}

type Conductor struct {
	client     llm.Client
	profile    *profile.Profile
	phase      Phase
	history    []Exchange
	questions  *QuestionBank
	transcript *Transcript
	phaseCount int
	pending    string
}

func NewConductor(client llm.Client, p *profile.Profile) *Conductor {
	return &Conductor{
		client:     client,
		profile:    p,
		phase:      PhaseIntro,
		questions:  NewQuestionBank(),
		transcript: NewTranscript(),
	}
}

func (c *Conductor) NextQuestion(ctx context.Context) (string, error) {
	if c.IsComplete() {
		return "", nil
	}

	c.maybeAdvancePhase()

	var question string
	var err error

	if len(c.history) == 0 {
		question = c.questions.SeedQuestion(c.phase)
	} else {
		question, err = c.generateQuestion(ctx)
		if err != nil {
			question = c.questions.SeedQuestion(c.phase)
		}
	}

	c.pending = question
	return question, nil
}

func (c *Conductor) SubmitAnswer(answer string) {
	ex := Exchange{
		Question: c.pending,
		Answer:   answer,
		Phase:    c.phase,
	}
	c.history = append(c.history, ex)
	c.transcript.Add(ex)
	c.phaseCount++
	c.pending = ""
}

func (c *Conductor) IsComplete() bool {
	total := len(c.history)
	if total >= maxQuestions {
		return true
	}
	if c.phase == PhaseWrapUp && c.phaseCount >= 2 && total >= minQuestions {
		return true
	}
	return false
}

func (c *Conductor) Transcript() *Transcript {
	c.transcript.Duration = time.Since(c.transcript.StartedAt)
	c.transcript.Complete = c.IsComplete()
	return c.transcript
}

func (c *Conductor) Phase() Phase {
	return c.phase
}

func (c *Conductor) QuestionCount() int {
	return len(c.history)
}

func (c *Conductor) LoadHistory(exchanges []Exchange) {
	c.history = exchanges
	for _, e := range exchanges {
		c.transcript.Add(e)
	}
	if len(exchanges) > 0 {
		c.phase = exchanges[len(exchanges)-1].Phase
		c.phaseCount = 0
		for _, e := range exchanges {
			if e.Phase == c.phase {
				c.phaseCount++
			}
		}
	}
}

func (c *Conductor) maybeAdvancePhase() {
	if c.phaseCount < questionsPerPhase {
		return
	}
	if c.phase >= PhaseWrapUp {
		return
	}

	remaining := maxQuestions - len(c.history)
	phasesLeft := int(PhaseWrapUp) - int(c.phase)
	if phasesLeft > 0 && remaining <= phasesLeft {
		c.phase = PhaseWrapUp
	} else {
		c.phase++
	}
	c.phaseCount = 0
}

func (c *Conductor) generateQuestion(ctx context.Context) (string, error) {
	systemPrompt := c.questions.SystemPrompt(c.phase, c.profile, c.history)

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: "Generate the next interview question."},
	}

	return c.client.Complete(ctx, messages, llm.CompletionOptions{
		Temperature: 0.8,
		MaxTokens:   200,
	})
}
