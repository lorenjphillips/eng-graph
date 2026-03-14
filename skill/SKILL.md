# eng-graph: Engineer Profile Generator

Generate reviewer profiles from real engineering communication data. You are the LLM — eng-graph is your data pipeline.

## Installation

```bash
go install github.com/lorenjphillips/eng-graph@latest
```

If Go is not installed, download the binary from https://github.com/lorenjphillips/eng-graph/releases

## Quick Start (One Command)

From inside any git repo with a GitHub remote:

```bash
eng-graph pipeline <name> --since 2025-01-01
```

This auto-detects the GitHub repo and user from `gh` CLI and git remote, ingests PR reviews/comments/PRs, and dumps all data as JSON to stdout. Status goes to stderr.

Override repo or user:

```bash
eng-graph pipeline emily --repos org/repo1,org/repo2 --user emilyz --since 2024-06-01
```

## What You Get

The JSON output contains an array of DataPoint objects:

```json
{
  "id": "github:org/repo:pr_review_comment:123456",
  "source": "github",
  "kind": "pr_review_comment",
  "author": "emilyz",
  "body": "This needs a nil check before dereferencing...",
  "timestamp": "2025-02-15T10:30:00Z",
  "context": {
    "repo": "org/repo",
    "pr_number": "42",
    "pr_title": "Add user auth",
    "file": "auth/handler.go",
    "diff_hunk": "@@ -10,6 +10,8 @@ func handleAuth..."
  }
}
```

Kinds: `pr_review`, `pr_review_comment`, `pr`

## Your Job: Analyze and Build the Profile

After getting the dump, analyze the data yourself. Look for:

1. **Philosophy** — What does this engineer value most? (correctness, readability, performance, simplicity)
2. **Review priorities** — What do they comment on repeatedly? (error handling, naming, tests, architecture)
3. **Approval criteria** — When do they approve vs request changes?
4. **Pet peeves** — What triggers strong reactions?
5. **Code patterns** — Specific patterns they enforce (nil checks, error wrapping, naming conventions)
6. **Communication style** — How do they phrase feedback? (direct, diplomatic, Socratic)
7. **Praise patterns** — What do they call out as good work?

## Output: Tiered Profile Structure

Generate four markdown files:

### Tier 1: persona.md (always loaded)
```markdown
# Emily Zhang - Engineering Persona

## Philosophy
[2-3 sentences about their core engineering values]

## Priorities
1. [Most important review focus]
2. [Second priority]
3. [etc.]

## Approval Criteria
[When they approve, when they request changes]

## Pet Peeves
- [Things that trigger strong feedback]
```

### Tier 2: review-patterns.md (loaded when reviewing specific file types)
```markdown
# Review Patterns

## Error Handling
**Trigger:** *.go, *.py
- [Specific patterns they enforce]

## Naming
**Trigger:** *
- [Naming conventions they care about]
```

### Tier 3: codebase-rules.md (loaded for specific codebases)
```markdown
# Codebase Rules
- [Specific wrappers, base classes, or patterns for their codebase]
```

### Tier 4: voice.md (loaded for tone matching)
```markdown
# Communication Voice

## Praise
- [How they express approval]

## Feedback
- [How they phrase criticism]

## Recurring Phrases
- [Phrases they use often]
```

## Step-by-Step Commands

If you need more control than `pipeline`, use individual commands:

```bash
eng-graph check --json              # What CLIs are available?
eng-graph init --auto               # Init + auto-detect GitHub
eng-graph profile create emily      # Create profile
eng-graph ingest emily --since 2025-01-01  # Ingest data
eng-graph dump emily                # Export as JSON
eng-graph dump emily --limit 100    # First 100 data points
```

## Interview (Optional)

To enrich the profile with the engineer's own input, conduct the interview yourself conversationally. Ask about:

1. What they value most in code review
2. Their top 3 review priorities
3. Common patterns they enforce
4. What makes them approve vs request changes
5. How they prefer to give feedback

Save the answers as a transcript:

```bash
eng-graph interview emily --transcript answers.json
```

## Prerequisites

- `gh` CLI installed and authenticated (`gh auth login`)
- Git repo with a GitHub remote (for auto-detection)
- Go 1.23+ (for `go install`)
