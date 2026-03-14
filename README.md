# eng-graph

Generate engineer reviewer profiles from real communication data. Ingest code reviews, PR comments, design discussions, and Slack messages, then synthesize structured profiles that any AI coding assistant can use to review code in a specific engineer's voice and style.

Works with Claude Code, Cursor, GitHub Copilot, Codex, and any LLM-powered tool that accepts system prompts.

## Why eng-graph

AI code review tools give generic feedback. Real engineers have specific priorities, pet peeves, naming preferences, and communication patterns built from years of experience. eng-graph captures those patterns from an engineer's actual communication history and generates tiered profiles that make AI reviews sound like the real person.

**What it does:**

- Connects to GitHub, GitLab, Notion, Google Docs, Confluence, Obsidian, and Quip
- Pulls PR reviews, review comments, PR descriptions, doc comments, and messages
- Runs an optional AI-driven interview to fill gaps in the data
- Uses any OpenAI-compatible LLM to analyze patterns and synthesize a profile
- Outputs agent-agnostic tiered markdown files ready to paste into any AI tool

## Installation

### From source

```bash
go install github.com/lorenjphillips/eng-graph@latest
```

### From binary

Download the latest release from the [releases page](https://github.com/lorenjphillips/eng-graph/releases).

### Build from source

```bash
git clone https://github.com/lorenjphillips/eng-graph.git
cd eng-graph
make build
./bin/eng-graph version
```

## Quick Start

```bash
# Initialize eng-graph in your project
eng-graph init

# Connect a GitHub data source
eng-graph source add github

# Create a profile for the engineer
eng-graph profile create jane-doe

# Ingest PR review data from the last year
eng-graph ingest jane-doe --since 2024-01-01

# Set your LLM API key (works with any OpenAI-compatible endpoint)
export ENG_GRAPH_LLM_API_KEY=sk-...

# Optionally run an AI-driven interview to enrich the profile
eng-graph interview jane-doe

# Build the profile from ingested data
eng-graph build jane-doe

# Review a PR in the engineer's voice
eng-graph review https://github.com/org/repo/pull/123 --profile jane-doe
```

## How It Works

eng-graph follows a pipeline: connect data sources, ingest raw communication, optionally interview the engineer, analyze patterns with an LLM, and render structured markdown.

```
Sources --> Adapters --> DataPoints --> Dedup --> SQLite     (no LLM)
                                                   |
                              Chunker --> Analyzer (LLM) --> Patterns
                                                   |
                    Patterns + Interview --> Synthesizer (LLM) --> Profile
                                                   |
                              Profile --> Renderer (templates) --> Tiered Markdown
```

The output is four tiered markdown files designed for on-demand loading:

```
.eng-graph/profiles/<name>/output/
  persona.md                  # Tier 1: Always loaded
  references/
    review-patterns.md        # Tier 2: Loaded when relevant files match
    codebase-rules.md         # Tier 3: Loaded for codebase-specific reviews
    voice.md                  # Tier 4: Loaded for tone-accurate communication
```

**Tier 1 -- Persona Core.** Engineering philosophy, ranked priorities, approval criteria, and pet peeves. Always included in the system prompt.

**Tier 2 -- Review Patterns.** File-glob-triggered review categories with specific patterns and example comments. Loaded when the PR touches matching files.

**Tier 3 -- Codebase Rules.** Required wrappers, base class conventions, naming rules, and code patterns the engineer enforces. Loaded for codebase-specific context.

**Tier 4 -- Communication Voice.** Exact praise phrases, feedback phrasing, recurring expressions, and tone description. Loaded when the review should match the engineer's voice.

## Commands

| Command | Description |
|---|---|
| `eng-graph init` | Create `.eng-graph/` directory with default configuration |
| `eng-graph source add <kind>` | Add a data source with guided prompts |
| `eng-graph source list` | List all configured data sources |
| `eng-graph source test <name>` | Verify connectivity to a data source |
| `eng-graph profile create <name>` | Create a new engineer profile |
| `eng-graph profile list` | List all profiles |
| `eng-graph ingest <profile>` | Pull data from sources into SQLite storage |
| `eng-graph interview <profile>` | Run an AI-driven interview in a terminal UI |
| `eng-graph build <profile>` | Analyze data and generate tiered markdown |
| `eng-graph review <pr-ref>` | Review a PR using a profile |
| `eng-graph export <profile>` | Export profile to a directory |
| `eng-graph version` | Print version |

### Flags

**Global:** `--dir` sets the working directory (default: current directory).

**ingest:** `--source` filters by source name. `--since` filters by date (format: `2024-01-01`). `--author` overrides the author filter.

**interview:** `--resume` continues a previous interview session. `--transcript` imports a transcript from a file.

**build:** `--tier` builds a specific tier (1-4, default: all). `--output` sets the output directory.

**review:** `--profile` selects the profile (default: `active_profile` from config). `--format` sets output format (`text`, `json`, or `gh`).

**export:** `--format` sets the export format (`md` or `json`). `--output` sets the target directory.

## Configuration

Running `eng-graph init` creates `.eng-graph/config.yaml`:

```yaml
llm:
  base_url: https://api.openai.com/v1
  api_key_env: ENG_GRAPH_LLM_API_KEY
  model: gpt-4o

sources:
  - name: work-github
    kind: github
    config:
      token_env: GITHUB_TOKEN
      user: janedoe
      repos:
        - myorg/backend
        - myorg/frontend

active_profile: jane-doe
```

### LLM Configuration

eng-graph works with any OpenAI-compatible API. Set `base_url` to your provider:

| Provider | base_url |
|---|---|
| OpenAI | `https://api.openai.com/v1` |
| Anthropic (via OpenRouter) | `https://openrouter.ai/api/v1` |
| Ollama (local) | `http://localhost:11434/v1` |
| Azure OpenAI | `https://{resource}.openai.azure.com/openai/deployments/{model}/v1` |

API keys are referenced by environment variable name (`api_key_env`), never stored in config files.

### Supported Data Sources

| Source | Kind | Required Config |
|---|---|---|
| GitHub | `github` | `token_env`, `repos`, `user` |
| GitLab | `gitlab` | `token_env`, `project` |
| Notion | `notion` | `token_env`, `database_ids` |
| Obsidian | `obsidian` | `vault_path` (local, no API key) |
| Google Docs | `gdocs` | `credentials_env`, `document_ids` |
| Confluence | `confluence` | `base_url`, `token_env`, `space_keys` |
| Quip | `quip` | `token_env`, `folder_ids` |

GitHub is fully implemented. Other adapters accept configuration and will be completed in upcoming releases.

## Using Profiles with AI Tools

### Claude Code

Add the persona to your project rules:

```bash
eng-graph export jane-doe --output .
cat persona.md >> CLAUDE.md
```

Or create a Claude Code skill from the full profile:

```bash
mkdir -p ~/.claude/skills/jane-doe
eng-graph export jane-doe --output ~/.claude/skills/jane-doe/
```

### Cursor

Add profile files to Cursor rules:

```bash
eng-graph export jane-doe --output .cursor/rules/
```

### GitHub Copilot

Set as Copilot instructions:

```bash
eng-graph export jane-doe --output .github/
mv .github/persona.md .github/copilot-instructions.md
```

### Programmatic Access

Export as JSON for custom integrations:

```bash
eng-graph export jane-doe --format json --output .
```

## Project Structure

```
eng-graph/
  main.go
  cmd/                          # CLI commands (thin Cobra wrappers)
  internal/
    config/                     # YAML configuration
    source/                     # Source adapter interface and registry
      github/                   # GitHub adapter (PR reviews, comments, descriptions)
      gitlab/ notion/ ...       # Additional adapters
    storage/                    # SQLite data point storage (CGo-free)
    ingest/                     # Ingestion pipeline with deduplication
    llm/                        # OpenAI-compatible LLM client
    builder/                    # LLM analysis, synthesis, and markdown rendering
      templates/                # Go text/template files for tiered output
    interview/                  # AI-driven interview with Bubbletea TUI
    review/                     # PR review engine with tier-aware prompts
```

## Requirements

- Go 1.23 or later
- A GitHub personal access token (for GitHub data source)
- An API key for any OpenAI-compatible LLM provider (for build, interview, and review commands)

## License

MIT License. See [LICENSE](LICENSE) for details.
