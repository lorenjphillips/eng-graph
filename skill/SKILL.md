# eng-graph

Use when the user says "build a profile", "engineer profile", "reviewer profile", "eng-graph", "ingest PRs", "review style", or wants to generate a reviewer persona from someone's GitHub activity.

## Setup

First check if eng-graph is installed:

```bash
which eng-graph
```

If not found, install it:

```bash
go install github.com/lorenjphillips/eng-graph@latest
```

If Go is not available, ask the user to install eng-graph from https://github.com/lorenjphillips/eng-graph/releases

Then verify GitHub CLI access:

```bash
eng-graph check --json
```

If `gh` is not authenticated, tell the user to run `gh auth login`.

## Workflow

When the user asks to build a profile (e.g. "build a profile for Emily from her PRs"):

### Step 1: Gather info from the user

Ask the user:
- Whose profile? (GitHub username)
- Which repos? (or auto-detect from current repo)
- How far back? (default: 6 months)

### Step 2: Run the pipeline

From inside a git repo with a GitHub remote:

```bash
eng-graph pipeline <name> --user <github-username> --since <YYYY-MM-DD>
```

To target specific repos instead of auto-detecting:

```bash
eng-graph pipeline <name> --user <github-username> --repos org/repo1,org/repo2 --since <YYYY-MM-DD>
```

This outputs JSON to stdout (status to stderr). Capture the JSON output.

### Step 3: Analyze the data

The JSON contains an array of DataPoint objects:

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

Analyze ALL the data points yourself. Look for:

1. **Philosophy** — What does this engineer value most?
2. **Review priorities** — What do they comment on repeatedly?
3. **Approval criteria** — When do they approve vs request changes?
4. **Pet peeves** — What triggers strong reactions?
5. **Code patterns** — Specific patterns they enforce
6. **Communication style** — How do they phrase feedback?
7. **Praise patterns** — What do they call out as good work?

### Step 4: Generate the profile

Write four tiered markdown files:

**Tier 1: persona.md** (always loaded — the engineer's core identity)
```markdown
# <Name> - Engineering Persona

## Philosophy
[2-3 sentences about their core engineering values, derived from patterns in their reviews]

## Priorities
1. [Most important review focus, with evidence]
2. [Second priority]
3. [etc.]

## Approval Criteria
[When they approve, when they request changes — based on actual review data]

## Pet Peeves
- [Things that consistently trigger strong feedback]
```

**Tier 2: review-patterns.md** (loaded when reviewing matching file types)
```markdown
# Review Patterns

## [Category Name]
**Trigger:** [glob patterns like *.go, *.py]
- [Specific patterns they enforce, with examples from their actual comments]

## [Another Category]
**Trigger:** [globs]
- [Patterns]
```

**Tier 3: codebase-rules.md** (loaded for specific codebases they work in)
```markdown
# Codebase Rules
- [Specific wrappers, base classes, naming conventions they enforce]
- [Internal patterns unique to their codebase]
```

**Tier 4: voice.md** (loaded for tone matching)
```markdown
# Communication Voice

## Praise
- [Direct quotes or patterns of how they express approval]

## Feedback
- [How they phrase criticism — direct, Socratic, diplomatic]

## Recurring Phrases
- [Phrases they use often, pulled directly from their comments]
```

Ask the user where to save the profile files.

## Individual Commands

For more control, use these separately:

```bash
eng-graph check --json                                    # Report CLI availability
eng-graph init --auto                                     # Init + auto-detect GitHub
eng-graph profile create <name>                           # Create profile
eng-graph ingest <name> --user <user> --since <date>      # Ingest data
eng-graph dump <name>                                     # Export as JSON
eng-graph dump <name> --limit 100                         # Limit output
```
