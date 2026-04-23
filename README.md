# qwilr-cli

A command-line interface for the [Qwilr API](https://docs.qwilr.com/api-reference/). Manage pages, templates, blocks, and webhooks from your terminal.

## Install

```bash
go install github.com/paulmeller/qwilr-cli@latest
```

Or build from source:

```bash
git clone https://github.com/paulmeller/qwilr-cli.git
cd qwilr-cli
go build -o qwilr .
```

## Using with Claude Code

Install the CLI, then add it as an MCP tool so Claude can manage your Qwilr pages directly:

```bash
# Install
go install github.com/paulmeller/qwilr-cli@latest

# Set your API token
export QWILR_API_TOKEN=your_token_here

# Add as a Claude Code slash command by putting this in your project's CLAUDE.md:
```

```markdown
## Tools

You have access to the `qwilr` CLI for managing Qwilr pages. Run it via bash:

- `qwilr pages list --json` — list all pages
- `qwilr pages create --from page.yaml` — create a page from a YAML spec
- `qwilr pages get <id> --json` — get page details
- `qwilr templates list --json` — list available templates
- `qwilr webhooks create --url <url> --event pageAccepted` — subscribe to events

Always use `--json` for structured output you can parse.
```

Or configure it as a [custom slash command](https://docs.anthropic.com/en/docs/claude-code/slash-commands) by creating `.claude/commands/qwilr.md`:

```markdown
List my Qwilr pages and templates, then ask what I'd like to do.
Use the `qwilr` CLI with `--json` for all commands.
```

Then use it in a session with `/project:qwilr`.

## Setup

Configure your API token (requires a Qwilr Enterprise account):

```bash
qwilr configure
```

Or set it via environment variable:

```bash
export QWILR_API_TOKEN=your_token_here
```

Config is stored at `~/.config/qwilr/config.yaml` with `0600` permissions. The env var takes precedence over the config file.

## Usage

### Pages

```bash
qwilr pages list
qwilr pages get <pageId>
qwilr pages create --name "Proposal" --template "sales-v2" --publish
qwilr pages create --from page.yaml
qwilr pages update <pageId> --name "New Name"
qwilr pages delete <pageId>
qwilr pages variables <pageId> --set customer=Acme --set value=50000
```

Create pages from a YAML file:

```yaml
# page.yaml
name: "Q3 Proposal for Acme Corp"
template: "sales-proposal-v2"
published: true
variables:
  customer_name: "Acme Corp"
  deal_value: "$50,000"
```

Flags override file values when both are provided:

```bash
qwilr pages create --from page.yaml --name "Override Name"
```

### Templates

```bash
qwilr templates list
```

### Blocks

```bash
qwilr blocks list
```

### Webhooks

```bash
qwilr webhooks list
qwilr webhooks create --url https://example.com/hook --event pageAccepted
qwilr webhooks delete <webhookId>
```

Supported events: `pageAccepted`, `pagePreviewAccepted`, `pageViewed`, `pageFirstViewed`

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format (default: human-friendly tables) |
| `--timeout` | API request timeout (default: 30s) |

```bash
# Pipe JSON output to jq
qwilr pages list --json | jq '.[].id'

# Custom timeout
qwilr pages list --timeout 10s
```

## Shell Completions

```bash
# Bash
source <(qwilr completion bash)

# Zsh
qwilr completion zsh > "${fpath[1]}/_qwilr"

# Fish
qwilr completion fish | source
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (API failure, network issue) |
| 2 | Validation error (missing token, invalid input) |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `QWILR_API_TOKEN` | API token (overrides config file) |
| `NO_COLOR` | Disable colored output when set |

## License

MIT
