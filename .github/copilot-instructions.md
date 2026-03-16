# Copilot Instructions for uictl

## Project Overview

**uictl** is a command-line interface for managing Ubiquiti UniFi network devices and controllers. It wraps the UniFi Controller API to provide a fast, scriptable CLI experience for network administrators. The design philosophy follows [clig.dev](https://clig.dev) and takes inspiration from `gh`, `kubectl`, and `aws` — CLIs that feel like natural extensions of the terminal.

**A core differentiator of uictl is first-class LLM/agent compatibility.** It is designed to be used seamlessly by AI coding agents (Copilot CLI, Claude Code, Gemini CLI, etc.) as well as human operators. Every design decision should consider both audiences: humans get beautiful, discoverable output; agents get structured, minimal, parseable output with runtime introspection and clear safety rails.

## CLI Name & Command Grammar

The binary/command name is `uictl`. Commands follow a **resource action** (noun-verb) pattern with flags over positional arguments:

```
uictl <resource> <action> [flags]
```

Examples:
- `uictl device list`
- `uictl device restart <device-id>`
- `uictl client list --site default`
- `uictl site list --json`
- `uictl network list --output table`
- `uictl login`
- `uictl api get /s/default/stat/device` (raw API passthrough)

Resources are **singular nouns** (e.g., `uictl device`, not `uictl devices`). Actions use clear, unambiguous verbs: `list`, `get`, `create`, `delete`, `restart`, `adopt`. Avoid overlapping verbs like `update` vs `upgrade`. Support both `-h` and `--help` on every command and subcommand.

Smart ID resolution: accept MAC addresses, device names, or internal IDs interchangeably — don't force users or agents to look up opaque numeric IDs.

For complex or nested input, support a **`--json-input` flag** that accepts a full JSON payload as the request body. Keep human-friendly flags (`--name`, `--vlan`) as well, but document that agents should prefer `--json-input` to avoid flag hallucination:
```
uictl network create --json-input '{"name": "IoT", "vlan_id": 30, "purpose": "corporate"}'
```

## Key Concepts

- **Controller**: The UniFi Controller (or UniFi OS Console) that manages the network.
- **Site**: UniFi organizes devices under sites. The default site is named `default`.
- **Device**: Any UniFi hardware (APs, switches, gateways, cameras, etc.).
- **Client**: A device connected to the network (wired or wireless).

## Discoverability

- Running `uictl` with no arguments prints a concise overview with grouped subcommands and usage examples.
- `--help` on every command shows **examples first**, then flags and descriptions.
- Typos should trigger "did you mean …?" suggestions.
- Ship shell completion for bash, zsh, and fish (`uictl completion <shell>`). Include dynamic completion where possible (e.g., tab-completing device names/IDs from the API).
- **`uictl schema <resource>.<action>`** — Runtime introspection command that returns a JSON schema for any command, including parameters, types, required fields, and copy-pasteable examples. This is the primary entry point for LLM agents discovering how to use a command (agents call `schema` before `--help`).
- **`uictl skills`** — Dumps concise, agent-optimized usage instructions (invariants, best practices, common patterns) for LLM agents to internalize in a single read.

## Output & Formatting

- **Default (TTY detected)**: Human-readable colored, aligned tables. Use spinners/progress indicators for any network call taking longer than ~300ms.
- **Non-TTY (piped/agent)**: Automatically switch to JSON output. This is critical for LLM agents — they should never receive colored table output by default.
- **`--json`**: Full JSON output, explicitly requested. For list commands, use NDJSON (one JSON object per line) to support streaming and reduce memory usage.
- **`--csv`**: CSV output for spreadsheets and simple parsing.
- **`--output <format>`**: Alternative long-form flag accepting `table`, `json`, `csv`, `ndjson`.
- **`--fields <field1,field2,...>`**: Field mask to select only specific fields in the response. Reduces token usage when agents only need `id,name,status` instead of the full object. This flag works with all output formats.
- **`--no-color`**: Force disable colors. Also respect the `NO_COLOR` environment variable.
- **`--quiet` / `-q`**: Suppress non-essential output.
- **`UICTL_OUTPUT_FORMAT`** env var: Agents/users can set this to `json` globally so every command returns structured output without per-command flags.
- **stdout** is for primary data output only. Logs, progress, errors, and prompts go to **stderr**.

## Feedback & Safety

- Every mutating action prints a clear confirmation: "Restarted device <name> (abc123) in 1.2s".
- Suggest next steps where helpful: "Run `uictl device get <id>` to check status."
- Destructive actions (delete, factory-reset, forget) **prompt for confirmation** unless `--force` or `--yes` is passed. Agents must always pass `--yes` explicitly — never present interactive prompts in non-TTY mode.
- **`--dry-run`** is supported on **every mutating command**. Agent guidance should recommend `--dry-run` first, then the real command after user confirmation.
- Errors are human-readable and actionable: explain what went wrong and what the user can do about it. Never show raw stack traces.
- **Agent-native error output**: All errors include a structured JSON object on stderr (when `--json` or non-TTY) with `code`, `message`, and `guidance` fields. The `guidance` field contains explicit next-step instructions for LLM agents:
  ```json
  {
    "code": "AUTH_EXPIRED",
    "message": "Session token has expired",
    "guidance": "Run `uictl login --profile <profile>` to re-authenticate, then retry the command."
  }
  ```

## Configuration & Auth

Precedence (highest wins): **CLI flags → environment variables → project config → user config**.

- User config lives at `~/.config/uictl/config` (XDG-compliant). Support `UICTL_CONFIG` env var to override.
- Auth flow: `uictl login` prompts for controller URL + credentials, stores a session token. Support `--profile` for managing multiple controllers.
- Secrets should be stored via the OS keyring where available, falling back to the config file with restrictive permissions (0600). Never log or echo secrets.
- Key environment variables: `UICTL_HOST`, `UICTL_SITE`, `UICTL_PROFILE`, `UICTL_NO_COLOR`, `UICTL_DEBUG`.
- `uictl config` subcommand for viewing/setting configuration.

## Scriptability & Robustness

- Commands are idempotent where possible.
- Meaningful exit codes: `0` = success, `1` = general error, `2` = auth/permission error, `3` = not found, `4` = conflict/validation error.
- Support reading input from stdin where it makes sense (e.g., `cat device-ids.txt | uictl device restart -`).
- Automatic retry with exponential backoff for transient API errors (429, 5xx).
- Handle API pagination transparently — return all results by default, or support `--limit` / `--page`.
- Respect `PAGER` environment variable for long output.

## API Notes

- The UniFi Controller API is REST-based and not officially documented by Ubiquiti.
- Authentication is session/cookie-based (login → receive cookie → use cookie for subsequent requests).
- API base path is typically `https://<controller>:8443/api/` (classic controller) or `https://<console>/proxy/network/api/` (UniFi OS).
- All API responses are JSON.
- Provide a raw `uictl api <method> <path> [--data <json>]` passthrough command for endpoints not yet wrapped by dedicated subcommands.

## Architecture & Code Conventions

- **Three-layer separation**: CLI parsing layer → API client library → output/formatting layer. Each layer is independently testable.
- The API client should be a standalone library/package (not tangled with CLI code) so it can be reused or tested in isolation.
- Write tests for: API client functions, output formatting, argument parsing/validation.
- Detect TTY in the output layer to decide between rich and plain rendering.
- Support a `--verbose` / `-v` global flag for debug-level logging (request URLs, timings, response codes) to stderr.
- Support a `--debug` flag (or `UICTL_DEBUG=1`) that additionally logs full request/response bodies.

## Distribution & Polish

- Aim for a single static binary where the language supports it.
- Include a `uictl version` command (with `--json` support) and consider an opt-in update checker.
- Ship man pages or markdown docs alongside the binary.

## LLM & Agent Integration

This is a first-class design pillar, not an afterthought. uictl should be the kind of CLI that LLM agents reach for naturally.

### Agent Workflow Assumptions

LLM agents using uictl will typically:
1. Run `uictl schema <resource>.<action>` or `uictl skills` to learn the command surface.
2. Use `--fields` to request only the data they need (minimizing token usage).
3. Use `--dry-run` before any mutating command, present the preview to the user, then execute.
4. Pass `--yes` explicitly on confirmed destructive actions.
5. Parse JSON from stdout; read guidance from structured errors on stderr.

### Design Rules for Agent Compatibility

- **No interactive prompts in non-TTY mode.** If stdin is not a TTY and `--yes` is not passed, fail with a clear error rather than hanging.
- **Deterministic output.** Same input → same output structure. Never add random tips, MOTD banners, or update nags to stdout.
- **Minimal output by default.** Agents don't need 40-field objects when 5 will do. `--fields` is the primary mechanism, but default JSON output should already be reasonably concise.
- **Input validation with clear rejection.** Sanitize and validate all inputs. Reject control characters, path traversal (`..`), and malformed IDs with a structured error explaining what's wrong and what's expected.
- **Smart ID resolution everywhere.** Accept MAC addresses (`aa:bb:cc:dd:ee:ff`), device names (`living-room-ap`), or internal UniFi IDs interchangeably.

### Agent Skills File

Ship a `SKILLS.md` file in the repository root that agents can read to internalize uictl best practices. This file should contain:

```yaml
---
tool: uictl
always:
  - use --fields on list/get commands to limit output
  - use --dry-run before any mutating command
  - use --json-input for complex create/update payloads
  - pass --yes for confirmed destructive actions
  - parse JSON output from stdout, errors from stderr
never:
  - rely on table-formatted output for parsing
  - omit --yes on destructive commands (will hang in non-TTY)
  - send unvalidated user input directly as device names or IDs
invariants:
  - resource names are singular nouns (device, client, network, site)
  - all timestamps are ISO 8601 / UTC
  - MAC addresses are lowercase colon-separated (aa:bb:cc:dd:ee:ff)
---
```

Followed by concise usage patterns and examples for common workflows.
