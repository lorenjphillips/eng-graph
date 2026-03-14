package review

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eng-graph/eng-graph/internal/llm"
	"github.com/eng-graph/eng-graph/internal/profile"
)

func BuildReviewPrompt(p *profile.Profile, diff *PRDiff) []llm.Message {
	var system strings.Builder

	// Tier 1: core persona (always included)
	system.WriteString("You are a senior engineer performing a code review. ")
	system.WriteString("Your review philosophy: " + p.Tier1.Philosophy + "\n\n")

	if len(p.Tier1.Priorities) > 0 {
		system.WriteString("Your priorities (in order):\n")
		for _, pri := range p.Tier1.Priorities {
			system.WriteString("- " + pri + "\n")
		}
		system.WriteString("\n")
	}

	if p.Tier1.ApprovalCriteria != "" {
		system.WriteString("Approval criteria: " + p.Tier1.ApprovalCriteria + "\n\n")
	}

	if len(p.Tier1.PetPeeves) > 0 {
		system.WriteString("Things you always flag:\n")
		for _, pp := range p.Tier1.PetPeeves {
			system.WriteString("- " + pp + "\n")
		}
		system.WriteString("\n")
	}

	// Tier 2: file-pattern-specific review patterns
	if matched := matchCategories(p.Tier2.Categories, diff.Files); len(matched) > 0 {
		system.WriteString("Relevant review patterns for files in this PR:\n")
		for _, cat := range matched {
			system.WriteString(fmt.Sprintf("\n[%s]\n", cat.Name))
			for _, pat := range cat.Patterns {
				system.WriteString("- " + pat + "\n")
			}
			if len(cat.Examples) > 0 {
				system.WriteString("Example comments you'd make:\n")
				for _, ex := range cat.Examples {
					system.WriteString("  > " + ex + "\n")
				}
			}
		}
		system.WriteString("\n")
	}

	// Tier 3: codebase rules
	if rules := formatRules(p.Tier3); rules != "" {
		system.WriteString("Codebase rules to enforce:\n" + rules + "\n")
	}

	// Tier 4: communication voice
	if p.Tier4.Tone != "" {
		system.WriteString("Communication style: " + p.Tier4.Tone + "\n")
	}
	if len(p.Tier4.PraisePhrases) > 0 {
		system.WriteString("When praising, use phrases like: " + strings.Join(p.Tier4.PraisePhrases, "; ") + "\n")
	}
	if len(p.Tier4.FeedbackPhrases) > 0 {
		system.WriteString("When giving feedback, use phrases like: " + strings.Join(p.Tier4.FeedbackPhrases, "; ") + "\n")
	}

	system.WriteString("\nReview the following PR. Provide actionable, specific feedback. Group by severity (blocking, suggestion, nit). Include file paths and line references where possible.")

	// User message: the diff
	var user strings.Builder
	user.WriteString(fmt.Sprintf("PR: %s\n", diff.Title))
	if diff.Description != "" {
		user.WriteString(fmt.Sprintf("\nDescription:\n%s\n", diff.Description))
	}
	user.WriteString(fmt.Sprintf("\nURL: %s\n\n", diff.URL))
	user.WriteString("Files changed:\n\n")

	for _, f := range diff.Files {
		user.WriteString(fmt.Sprintf("--- %s (%s) ---\n", f.Path, f.Status))
		if f.Patch != "" {
			user.WriteString(f.Patch)
			user.WriteString("\n\n")
		}
	}

	return []llm.Message{
		{Role: llm.RoleSystem, Content: system.String()},
		{Role: llm.RoleUser, Content: user.String()},
	}
}

func matchCategories(categories []profile.ReviewCategory, files []DiffFile) []profile.ReviewCategory {
	var matched []profile.ReviewCategory
	seen := map[string]bool{}

	for _, cat := range categories {
		if seen[cat.Name] {
			continue
		}
		for _, f := range files {
			if categoryMatchesFile(cat, f.Path) {
				matched = append(matched, cat)
				seen[cat.Name] = true
				break
			}
		}
	}
	return matched
}

func categoryMatchesFile(cat profile.ReviewCategory, path string) bool {
	for _, glob := range cat.TriggerGlobs {
		if strings.Contains(glob, "**") {
			prefix := strings.TrimSuffix(glob, "/**")
			prefix = strings.TrimSuffix(prefix, "**")
			if prefix == "" || strings.HasPrefix(path, prefix) {
				return true
			}
			continue
		}
		if ok, _ := filepath.Match(glob, path); ok {
			return true
		}
		if ok, _ := filepath.Match(glob, filepath.Base(path)); ok {
			return true
		}
	}
	return false
}

func formatRules(rules profile.CodebaseRules) string {
	var b strings.Builder
	writeRules := func(label string, rs []profile.CodeRule) {
		if len(rs) == 0 {
			return
		}
		b.WriteString(fmt.Sprintf("  %s:\n", label))
		for _, r := range rs {
			b.WriteString("    - " + r.Rule)
			if r.Example != "" {
				b.WriteString(" (e.g., " + r.Example + ")")
			}
			b.WriteString("\n")
		}
	}
	writeRules("Wrappers", rules.Wrappers)
	writeRules("Base classes", rules.BaseClasses)
	writeRules("Naming", rules.Naming)
	writeRules("Patterns", rules.Patterns)
	return b.String()
}
