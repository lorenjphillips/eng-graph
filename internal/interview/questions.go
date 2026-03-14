package interview

import (
	"fmt"
	"strings"

	"github.com/eng-graph/eng-graph/internal/profile"
)

type QuestionBank struct {
	seedQuestions map[Phase][]string
}

func NewQuestionBank() *QuestionBank {
	return &QuestionBank{
		seedQuestions: map[Phase][]string{
			PhaseIntro: {
				"Tell me about your current role and the kinds of systems you work on day to day.",
				"How long have you been doing code reviews, and how has your approach evolved?",
			},
			PhasePriorities: {
				"When you open a pull request, what's the first thing you look at?",
				"What would make you immediately request changes on a PR?",
				"How do you decide what feedback is worth leaving versus what to let slide?",
			},
			PhaseCodePatterns: {
				"Are there specific code patterns or idioms you always look for?",
				"What naming conventions do you feel strongly about?",
				"How do you approach reviewing error handling?",
			},
			PhaseArchitecture: {
				"How do you evaluate whether a PR's design fits the broader system?",
				"When do you push back on architecture versus accepting pragmatic solutions?",
			},
			PhaseErrors: {
				"What error handling patterns do you consider essential?",
				"How do you feel about defensive coding versus failing fast?",
			},
			PhaseDatabase: {
				"What do you look for when reviewing database-related changes?",
				"How do you evaluate migration safety in code review?",
			},
			PhaseCommunication: {
				"How would you describe your tone when giving critical feedback?",
				"Do you have any phrases or patterns you find yourself using repeatedly in reviews?",
				"How do you balance being thorough with being encouraging?",
			},
			PhaseWrapUp: {
				"Is there anything about your review style we haven't covered that you think is important?",
				"If you could give one piece of advice to someone whose code you review, what would it be?",
			},
		},
	}
}

func (q *QuestionBank) SeedQuestion(phase Phase) string {
	seeds := q.seedQuestions[phase]
	if len(seeds) == 0 {
		return "Tell me more about your engineering preferences."
	}
	return seeds[0]
}

func (q *QuestionBank) SystemPrompt(phase Phase, p *profile.Profile, history []Exchange) string {
	var b strings.Builder

	b.WriteString("You are conducting a conversational interview to understand an engineer's code review style, ")
	b.WriteString("engineering philosophy, and technical preferences. Your goal is to build a detailed profile ")
	b.WriteString("of how they think about code.\n\n")

	b.WriteString("Guidelines:\n")
	b.WriteString("- Ask exactly ONE question at a time\n")
	b.WriteString("- Be conversational and natural, not robotic\n")
	b.WriteString("- Reference their previous answers to build on what they've said\n")
	b.WriteString("- If they mention something interesting, follow up on it before moving on\n")
	b.WriteString("- Keep questions open-ended but specific enough to get actionable insights\n")
	b.WriteString("- Output ONLY the question text, nothing else\n\n")

	b.WriteString(fmt.Sprintf("Current phase: %s\n", phaseName(phase)))
	b.WriteString(fmt.Sprintf("Phase focus: %s\n\n", phaseDescription(phase)))

	if p != nil {
		writeProfileContext(&b, p)
	}

	if len(history) > 0 {
		writeHistory(&b, history)
	}

	seeds := q.seedQuestions[phase]
	if len(seeds) > 0 {
		b.WriteString("Example questions for this phase (use as inspiration, not verbatim):\n")
		for _, s := range seeds {
			b.WriteString(fmt.Sprintf("- %s\n", s))
		}
		b.WriteString("\n")
	}

	b.WriteString("Generate the next interview question.")
	return b.String()
}

func writeProfileContext(b *strings.Builder, p *profile.Profile) {
	b.WriteString("Known profile data (reference specific items to personalize questions):\n")

	if p.Tier1.Philosophy != "" {
		b.WriteString(fmt.Sprintf("- Philosophy: %s\n", p.Tier1.Philosophy))
	}
	if len(p.Tier1.Priorities) > 0 {
		b.WriteString(fmt.Sprintf("- Priorities: %s\n", strings.Join(p.Tier1.Priorities, ", ")))
	}
	if len(p.Tier1.PetPeeves) > 0 {
		b.WriteString(fmt.Sprintf("- Pet peeves: %s\n", strings.Join(p.Tier1.PetPeeves, ", ")))
	}
	for _, cat := range p.Tier2.Categories {
		if len(cat.Patterns) > 0 {
			b.WriteString(fmt.Sprintf("- Review pattern (%s): %s\n", cat.Name, strings.Join(cat.Patterns[:min(3, len(cat.Patterns))], "; ")))
		}
	}
	if len(p.Tier3.Naming) > 0 {
		b.WriteString("- Has strong naming conventions\n")
	}
	if len(p.Tier3.Wrappers) > 0 {
		b.WriteString("- Enforces wrapper usage patterns\n")
	}
	if p.Tier4.Tone != "" {
		b.WriteString(fmt.Sprintf("- Communication tone: %s\n", p.Tier4.Tone))
	}
	if len(p.Tier4.RecurringPhrases) > 0 {
		b.WriteString(fmt.Sprintf("- Recurring phrases: %s\n", strings.Join(p.Tier4.RecurringPhrases[:min(3, len(p.Tier4.RecurringPhrases))], "; ")))
	}
	b.WriteString("\n")
}

func writeHistory(b *strings.Builder, history []Exchange) {
	b.WriteString("Conversation so far:\n")
	for _, ex := range history {
		b.WriteString(fmt.Sprintf("Q: %s\n", ex.Question))
		b.WriteString(fmt.Sprintf("A: %s\n\n", ex.Answer))
	}
}

func phaseName(p Phase) string {
	names := map[Phase]string{
		PhaseIntro:         "Introduction",
		PhasePriorities:    "Review Priorities",
		PhaseCodePatterns:  "Code Patterns",
		PhaseArchitecture:  "Architecture",
		PhaseErrors:        "Error Handling",
		PhaseDatabase:      "Database",
		PhaseCommunication: "Communication Style",
		PhaseWrapUp:        "Wrap Up",
	}
	if n, ok := names[p]; ok {
		return n
	}
	return "Unknown"
}

func phaseDescription(p Phase) string {
	desc := map[Phase]string{
		PhaseIntro:         "Get to know the engineer and their background",
		PhasePriorities:    "Understand what they prioritize in code reviews",
		PhaseCodePatterns:  "Learn about specific code patterns and idioms they care about",
		PhaseArchitecture:  "Explore how they think about system design and architecture",
		PhaseErrors:        "Understand their approach to error handling and resilience",
		PhaseDatabase:      "Learn about database-related review patterns",
		PhaseCommunication: "Understand their communication style and tone in reviews",
		PhaseWrapUp:        "Capture anything we missed and get final thoughts",
	}
	if d, ok := desc[p]; ok {
		return d
	}
	return ""
}
