# uictl — UniFi CLI for Humans and Agents

This project includes `uictl`, a CLI for managing UniFi network controllers.
If `uictl` is installed and available in PATH, use it for all UniFi operations.

## Discovering uictl

```bash
# Check if available
which uictl

# Learn all commands (preferred for agents)
uictl schema

# Full skills reference
uictl skills
```

## Using uictl

Command pattern: `uictl <resource> <action> [flags]`

### Rules

- Use `--fields` on list/get commands to limit output and save tokens
- Use `--dry-run` before any mutating command, then confirm with the user
- Use `--json-input` for complex create/update payloads (avoids flag hallucination)
- Pass `--yes` on confirmed destructive actions (delete, remove)
- Parse JSON from stdout; errors go to stderr with a `guidance` field
- Non-TTY mode automatically outputs JSON — no need for `--json`

### Common Patterns

```bash
# List with field selection
uictl device list --fields id,name,model,state

# Introspect a command before using it
uictl schema network.create

# Create with JSON input
uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'

# Safe mutation: dry-run first, then execute
uictl network delete <id> --dry-run
uictl network delete <id> --yes

# Raw API for anything not yet wrapped
uictl api get /v1/info

# Classic API access (--raw bypasses /integration/ prefix)
uictl api get --raw /proxy/network/api/s/default/stat/device

# Port VLAN management
uictl device port list <device-id> --fields idx,name,nativeNetwork,speed,up
uictl device port set <device-id> 4 --network VLAN80_Home --dry-run
```

### Environment Setup

```bash
export UICTL_HOST=<controller-ip>
export UICTL_API_KEY=<api-key>
export UICTL_SITE=default
```

## Invariants

- Resource names are singular nouns: device, client, network, site
- All timestamps are ISO 8601 / UTC
- MAC addresses are lowercase colon-separated: aa:bb:cc:dd:ee:ff
- IDs are UUIDs
- `--site default` resolves the name to the correct UUID automatically

## Boundaries

- **Always do**: Use `--dry-run` before mutations, use `--fields` to minimize output
- **Ask the user**: Before any destructive action (delete, remove, factory-reset)
- **Never do**: Bypass `--yes` on destructive commands without user confirmation, send unvalidated input as device names or IDs
