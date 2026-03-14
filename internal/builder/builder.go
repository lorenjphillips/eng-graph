package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/eng-graph/eng-graph/internal/ingest"
	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
	"github.com/eng-graph/eng-graph/internal/source"
)

const batchTokenLimit = 80000

type Builder struct {
	client llm.Client
}

func NewBuilder(client llm.Client) *Builder {
	return &Builder{client: client}
}

func (b *Builder) Build(ctx context.Context, p *profile.Profile, datapoints []source.DataPoint, interviewTranscript json.RawMessage) error {
	chunks := ingest.ChunkDataPoints(datapoints, batchTokenLimit)
	fmt.Fprintf(os.Stderr, "analyzing %d data points in %d batches\n", len(datapoints), len(chunks))

	var allPatterns []analysisResult
	for i, chunk := range chunks {
		fmt.Fprintf(os.Stderr, "analyzing batch %d/%d (%d data points)\n", i+1, len(chunks), len(chunk))
		result, err := b.analyze(ctx, p.Name, chunk)
		if err != nil {
			return fmt.Errorf("analysis batch %d: %w", i+1, err)
		}
		allPatterns = append(allPatterns, result)
	}

	merged := mergeAnalyses(allPatterns)

	fmt.Fprintf(os.Stderr, "synthesizing profile\n")
	if err := b.synthesize(ctx, p, merged, interviewTranscript); err != nil {
		return fmt.Errorf("synthesis: %w", err)
	}

	p.DataSources = summarizeDataSources(datapoints)
	fmt.Fprintf(os.Stderr, "profile built successfully\n")
	return nil
}

type analysisResult struct {
	Philosophy    []string        `json:"philosophy"`
	Priorities    []string        `json:"priorities"`
	PetPeeves     []string        `json:"pet_peeves"`
	Categories    []categoryDraft `json:"categories"`
	Wrappers      []profile.CodeRule `json:"wrappers"`
	BaseClasses   []profile.CodeRule `json:"base_classes"`
	Naming        []profile.CodeRule `json:"naming"`
	Patterns      []profile.CodeRule `json:"patterns"`
	PraisePhrases []string        `json:"praise_phrases"`
	FeedbackPhrases []string      `json:"feedback_phrases"`
	RecurringPhrases []string     `json:"recurring_phrases"`
	Tone          string          `json:"tone"`
}

type categoryDraft struct {
	Name         string   `json:"name"`
	TriggerGlobs []string `json:"trigger_globs"`
	Patterns     []string `json:"patterns"`
	Examples     []string `json:"examples"`
}

func (b *Builder) analyze(ctx context.Context, name string, dps []source.DataPoint) (analysisResult, error) {
	dpJSON, err := json.Marshal(dps)
	if err != nil {
		return analysisResult{}, err
	}

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: analyzerSystemPrompt},
		{Role: llm.RoleUser, Content: fmt.Sprintf("Engineer: %s\n\nData points:\n%s", name, string(dpJSON))},
	}

	resp, err := b.client.Complete(ctx, messages, llm.CompletionOptions{
		Temperature: 0.2,
		MaxTokens:   8192,
		JSONMode:    true,
	})
	if err != nil {
		return analysisResult{}, err
	}

	var result analysisResult
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return analysisResult{}, fmt.Errorf("parsing analysis response: %w", err)
	}
	return result, nil
}

func (b *Builder) synthesize(ctx context.Context, p *profile.Profile, merged analysisResult, interview json.RawMessage) error {
	mergedJSON, err := json.Marshal(merged)
	if err != nil {
		return err
	}

	userContent := fmt.Sprintf("Engineer: %s\n\nMerged analysis:\n%s", p.Name, string(mergedJSON))
	if len(interview) > 0 {
		userContent += fmt.Sprintf("\n\nInterview transcript:\n%s", string(interview))
	}

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: synthesizerSystemPrompt},
		{Role: llm.RoleUser, Content: userContent},
	}

	resp, err := b.client.Complete(ctx, messages, llm.CompletionOptions{
		Temperature: 0.3,
		MaxTokens:   8192,
		JSONMode:    true,
	})
	if err != nil {
		return err
	}

	var synthesized struct {
		DisplayName      string                  `json:"display_name"`
		Tier1            profile.PersonaCore     `json:"tier1"`
		Tier2            profile.ReviewPatterns   `json:"tier2"`
		Tier3            profile.CodebaseRules    `json:"tier3"`
		Tier4            profile.CommunicationVoice `json:"tier4"`
	}
	if err := json.Unmarshal([]byte(resp), &synthesized); err != nil {
		return fmt.Errorf("parsing synthesis response: %w", err)
	}

	if synthesized.DisplayName != "" {
		p.DisplayName = synthesized.DisplayName
	}
	p.Tier1 = synthesized.Tier1
	p.Tier2 = synthesized.Tier2
	p.Tier3 = synthesized.Tier3
	p.Tier4 = synthesized.Tier4
	return nil
}

func mergeAnalyses(results []analysisResult) analysisResult {
	merged := analysisResult{}
	seen := struct {
		philosophy, priorities, peeves           map[string]bool
		praise, feedback, recurring              map[string]bool
	}{
		make(map[string]bool), make(map[string]bool), make(map[string]bool),
		make(map[string]bool), make(map[string]bool), make(map[string]bool),
	}

	for _, r := range results {
		merged.Philosophy = appendUnique(merged.Philosophy, r.Philosophy, seen.philosophy)
		merged.Priorities = appendUnique(merged.Priorities, r.Priorities, seen.priorities)
		merged.PetPeeves = appendUnique(merged.PetPeeves, r.PetPeeves, seen.peeves)
		merged.PraisePhrases = appendUnique(merged.PraisePhrases, r.PraisePhrases, seen.praise)
		merged.FeedbackPhrases = appendUnique(merged.FeedbackPhrases, r.FeedbackPhrases, seen.feedback)
		merged.RecurringPhrases = appendUnique(merged.RecurringPhrases, r.RecurringPhrases, seen.recurring)
		merged.Wrappers = append(merged.Wrappers, r.Wrappers...)
		merged.BaseClasses = append(merged.BaseClasses, r.BaseClasses...)
		merged.Naming = append(merged.Naming, r.Naming...)
		merged.Patterns = append(merged.Patterns, r.Patterns...)
		merged.Categories = mergeCategories(merged.Categories, r.Categories)

		if r.Tone != "" && merged.Tone == "" {
			merged.Tone = r.Tone
		} else if r.Tone != "" {
			merged.Tone = merged.Tone + "; " + r.Tone
		}
	}
	return merged
}

func appendUnique(dst, src []string, seen map[string]bool) []string {
	for _, s := range src {
		key := strings.ToLower(strings.TrimSpace(s))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		dst = append(dst, s)
	}
	return dst
}

func mergeCategories(dst, src []categoryDraft) []categoryDraft {
	byName := make(map[string]int)
	for i, c := range dst {
		byName[strings.ToLower(c.Name)] = i
	}
	for _, c := range src {
		key := strings.ToLower(c.Name)
		if idx, ok := byName[key]; ok {
			dst[idx].Patterns = append(dst[idx].Patterns, c.Patterns...)
			dst[idx].Examples = append(dst[idx].Examples, c.Examples...)
			for _, g := range c.TriggerGlobs {
				dst[idx].TriggerGlobs = append(dst[idx].TriggerGlobs, g)
			}
		} else {
			byName[key] = len(dst)
			dst = append(dst, c)
		}
	}
	return dst
}

func summarizeDataSources(dps []source.DataPoint) []profile.DataSourceSummary {
	type key struct{ source, kind string }
	counts := make(map[key]int)
	for _, dp := range dps {
		counts[key{dp.Source, string(dp.Kind)}]++
	}

	summaries := make([]profile.DataSourceSummary, 0, len(counts))
	for k, c := range counts {
		summaries = append(summaries, profile.DataSourceSummary{
			Source: k.source,
			Kind:   k.kind,
			Count:  c,
		})
	}
	return summaries
}

const analyzerSystemPrompt = `You are an engineering profile analyzer. Given a set of data points (code review comments, PR descriptions, messages) from a specific engineer, extract patterns about their review style, coding preferences, and communication voice.

Return a JSON object with these fields:
- philosophy: array of strings capturing the engineer's core review philosophy
- priorities: array of strings listing what they prioritize in code reviews
- pet_peeves: array of strings listing things that consistently bother them
- categories: array of {name, trigger_globs, patterns, examples} for review pattern categories
- wrappers: array of {rule, example} for wrapper/client usage rules they enforce
- base_classes: array of {rule, example} for base class/inheritance rules
- naming: array of {rule, example} for naming conventions they enforce
- patterns: array of {rule, example} for general code patterns they enforce
- praise_phrases: array of exact phrases they use when approving
- feedback_phrases: array of exact phrases they use when requesting changes
- recurring_phrases: array of phrases they repeat frequently
- tone: string describing their communication tone

Extract only what is clearly evidenced in the data. Do not invent patterns.`

const synthesizerSystemPrompt = `You are a profile synthesizer. Given merged analysis results from an engineer's code reviews and communications, produce a final, deduplicated, well-organized profile.

Return a JSON object with these fields:
- display_name: the engineer's preferred display name (infer from data, or use the name given)
- tier1: {philosophy (single coherent paragraph), priorities (ordered list, most important first), approval_criteria (concise paragraph), pet_peeves (list)}
- tier2: {categories: [{name, trigger_globs, patterns, examples}]} - deduplicated and organized
- tier3: {wrappers: [{rule, example}], base_classes: [{rule, example}], naming: [{rule, example}], patterns: [{rule, example}]} - deduplicated
- tier4: {tone (paragraph), praise_phrases (exact quotes), feedback_phrases (exact quotes), recurring_phrases (exact quotes)}

Deduplicate aggressively. Merge similar items. Keep the engineer's actual voice and phrasing in tier4. Make tier1 philosophy a single coherent narrative, not a list.`
