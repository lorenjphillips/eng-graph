package interview

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			PaddingBottom(1)

	phaseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)

	interviewerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")).
				Bold(true)

	interviewerBody = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			PaddingLeft(2).
			PaddingBottom(1)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true)

	userBody = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			PaddingLeft(2).
			PaddingBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

type chatMsg struct {
	role string
	text string
}

type model struct {
	ctx       context.Context
	conductor *Conductor
	messages  []chatMsg
	textarea  textarea.Model
	width     int
	height    int
	loading   bool
	err       error
	done      bool
	scroll    int
}

type questionMsg string
type errMsg struct{ err error }
type doneMsg struct{}

func RunInterview(ctx context.Context, conductor *Conductor) (*Transcript, error) {
	ta := textarea.New()
	ta.Placeholder = "Type your answer..."
	ta.Focus()
	ta.CharLimit = 2000
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	m := model{
		ctx:       ctx,
		conductor: conductor,
		textarea:  ta,
		loading:   true,
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := result.(*model)
	return final.conductor.Transcript(), final.err
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.fetchQuestion())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.done = true
			return m, tea.Quit

		case tea.KeyEnter:
			if m.loading || m.done {
				return m, nil
			}
			answer := strings.TrimSpace(m.textarea.Value())
			if answer == "" {
				return m, nil
			}

			m.messages = append(m.messages, chatMsg{role: "user", text: answer})
			m.conductor.SubmitAnswer(answer)
			m.textarea.Reset()
			m.scroll = 0

			if m.conductor.IsComplete() {
				m.done = true
				return m, tea.Quit
			}

			m.loading = true
			return m, m.fetchQuestion()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		return m, nil

	case questionMsg:
		m.loading = false
		m.messages = append(m.messages, chatMsg{role: "interviewer", text: string(msg)})
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, tea.Quit

	case doneMsg:
		m.done = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header
	progress := fmt.Sprintf("Question %d of ~20", m.conductor.QuestionCount()+1)
	phase := fmt.Sprintf("[%s]", phaseName(m.conductor.Phase()))
	header := headerStyle.Render("eng-graph interview") + "  " +
		phaseStyle.Render(phase) + "  " +
		statusStyle.Render(progress)
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render(strings.Repeat("─", min(m.width, 80))))
	b.WriteString("\n\n")

	// Chat messages
	chatContent := m.renderMessages()
	b.WriteString(chatContent)

	if m.loading {
		b.WriteString(interviewerStyle.Render("Interviewer"))
		b.WriteString("\n")
		b.WriteString(interviewerBody.Render("Thinking..."))
		b.WriteString("\n")
	}

	// Input area
	b.WriteString(dividerStyle.Render(strings.Repeat("─", min(m.width, 80))))
	b.WriteString("\n")
	if !m.done {
		b.WriteString(m.textarea.View())
	}
	b.WriteString("\n")
	b.WriteString(statusStyle.Render("Enter: submit  |  Ctrl+C: quit"))

	return b.String()
}

func (m *model) renderMessages() string {
	var b strings.Builder

	maxWidth := min(m.width-4, 78)

	// Show at most the last N messages that fit
	visible := m.messages
	if len(visible) > 20 {
		visible = visible[len(visible)-20:]
	}

	for _, msg := range visible {
		switch msg.role {
		case "interviewer":
			b.WriteString(interviewerStyle.Render("Interviewer"))
			b.WriteString("\n")
			b.WriteString(interviewerBody.Width(maxWidth).Render(msg.text))
			b.WriteString("\n")
		case "user":
			b.WriteString(userStyle.Render("You"))
			b.WriteString("\n")
			b.WriteString(userBody.Width(maxWidth).Render(msg.text))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *model) fetchQuestion() tea.Cmd {
	return func() tea.Msg {
		q, err := m.conductor.NextQuestion(m.ctx)
		if err != nil {
			return errMsg{err}
		}
		if q == "" {
			return doneMsg{}
		}
		return questionMsg(q)
	}
}
