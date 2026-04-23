# Qwilr CLI Design Spec

## Overview

A Go CLI for managing Qwilr resources (pages, templates, blocks, webhooks) via the Qwilr API. Built with Cobra + Viper. Compiles to a single binary with no runtime dependencies.

## API Target

- Base URL: `https://api.qwilr.com/v1`
- Auth: Bearer token via `Authorization` header

## Command Structure

```
qwilr
├── configure                    # Set up API token + preferences
├── pages
│   ├── list                     # List pages (--limit, --offset)
│   ├── create                   # Create page (--name, --template, --from file.yaml, --publish)
│   ├── get <pageId>             # Get page details
│   ├── update <pageId>          # Update page (--name, --content)
│   ├── delete <pageId>          # Delete page
│   └── variables <pageId>       # Set token/substitution values (--from file.yaml or --set key=value)
├── templates
│   └── list                     # List available templates
├── blocks
│   └── list                     # List saved blocks
├── webhooks
│   ├── list                     # List webhook subscriptions
│   ├── create                   # Create subscription (--url, --event)
│   └── delete <webhookId>       # Delete subscription
└── completion                   # Generate shell completions (bash/zsh/fish)
```

All commands support `--json` for machine-readable output. Human-friendly tables by default.

## Authentication & Configuration

Config file: `~/.config/qwilr/config.yaml`

```yaml
api_token: "qwilr_abc123..."
default_output: "text"  # or "json"
```

Precedence (highest to lowest):
1. `--json` flag (per-command override)
2. `QWILR_API_TOKEN` env var
3. Config file values

`qwilr configure` interactively prompts for the token and writes the config file with `0600` permissions.

## Project Layout

```
qwilr-cli/
├── main.go                     # Entry point
├── go.mod
├── cmd/
│   ├── root.go                 # Root command, global flags (--json)
│   ├── configure.go            # configure command
│   ├── pages.go                # pages subcommands
│   ├── templates.go            # templates list
│   ├── blocks.go               # blocks list
│   ├── webhooks.go             # webhooks subcommands
│   └── completion.go           # shell completion
├── internal/
│   ├── api/
│   │   └── client.go           # HTTP client, auth, base URL, error handling
│   ├── config/
│   │   └── config.go           # Load/save config, env var precedence
│   └── output/
│       └── output.go           # Table vs JSON formatting
└── README.md
```

- `internal/api/client.go` — single HTTP client struct wrapping `net/http`, handles auth header injection, JSON marshaling, error responses
- `internal/output/` — one place for formatting logic; commands call `output.Print(data, isJSON)` and don't worry about format
- One file per resource group in `cmd/`

## Page Creation from Files

YAML/JSON file defines the full page spec:

```yaml
name: "Q3 Proposal for Acme Corp"
template: "sales-proposal-v2"
published: true
variables:
  customer_name: "Acme Corp"
  deal_value: "$50,000"
  rep_name: "Jane Smith"
```

Usage:
```bash
# From file
qwilr pages create --from page.yaml

# From flags
qwilr pages create --name "Quick Proposal" --template "blank" --publish

# Hybrid — file as base, flags override
qwilr pages create --from page.yaml --name "Override Name"
```

Flag values override file values when both are provided.

## Output Formatting

Human-friendly (default):
```
ID              NAME                    STATUS    CREATED
pg_abc123       Q3 Proposal - Acme      live      2026-04-20
pg_def456       Onboarding Guide        draft     2026-04-18
```

Detail view:
```
Name:       Q3 Proposal - Acme
ID:         pg_abc123
Status:     live
Created:    2026-04-20
URL:        https://pages.qwilr.com/abc123
```

JSON mode (`--json`) outputs raw JSON arrays/objects.

Errors go to stderr in both modes. Table rendering via `text/tabwriter` (standard library).

## Error Handling

API client maps HTTP status codes to actionable messages:
- `401` -> "Authentication failed -- run `qwilr configure`"
- `404` -> "Resource not found: <id>"
- `429` -> "Rate limited -- retry after <n> seconds"
- `5xx` -> "Qwilr API error -- try again or check https://status.qwilr.com"

No automatic retries in v1. Surface the error, let the user retry.

## Dependencies

- `github.com/spf13/cobra` — command structure
- `github.com/spf13/viper` — config file + env var handling
- `text/tabwriter` (standard library) — table output
- `gopkg.in/yaml.v3` — YAML file parsing for `--from`

## Out of Scope (v1)

- Automatic retries / `--retry` flag
- Page sections endpoint (POST `/pages/{pageId}/sections`)
- Quote block data management
- Interactive page builder
- OAuth flow (API token only)
