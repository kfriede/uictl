# uictl

A command-line interface for managing Ubiquiti UniFi network devices and controllers. Built for both humans and LLM agents.

## Features

- **Full Network API coverage** — All 14 resource groups (73 operations) from the UniFi Network Integration API v1
- **Full Protect API coverage** — Cameras, lights, sensors, chimes, viewers, live views, NVR, alarm manager (35 operations)
- **LLM/agent-first design** — Auto-JSON in non-TTY, `--fields` for token-efficient output, `uictl schema` for runtime introspection, structured error guidance
- **Human-friendly output** — Colored tables when interactive, JSON/CSV/NDJSON when piped
- **Safety rails** — `--dry-run` on every mutation, confirmation prompts on destructive actions, `--yes` for automation
- **Secure credentials** — OS keyring (macOS Keychain, Linux secret-service, Windows Credential Manager) with config file fallback
- **Multi-profile** — Manage multiple controllers with named profiles

## Install

```bash
go install github.com/kfriede/uictl@latest
```

Or build from source:

```bash
git clone https://github.com/kfriede/uictl.git
cd uictl
go build -o uictl .
```

## Quick Start

```bash
# Authenticate with your controller
uictl login

# List sites
uictl site list

# List devices with specific fields
uictl device list --fields id,name,model,state

# Create a network
uictl network create --name "IoT" --vlan 30

# Or use JSON input (preferred for agents)
uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'

# Restart a device (with dry-run first)
uictl device restart <device-id> --dry-run
uictl device restart <device-id>

# Generate guest vouchers
uictl hotspot create --name "Day Pass" --duration 1440 --count 10

# Raw API passthrough
uictl api get /v1/info
```

## Commands

| Command | Description |
|---|---|
| `uictl login` | Authenticate with a UniFi controller |
| `uictl info` | Get controller application info |
| `uictl site list` | List local sites |
| `uictl device list\|get\|adopt\|remove\|restart\|stats\|pending` | Manage devices |
| `uictl client list\|get\|authorize\|unauthorize` | Manage connected clients |
| `uictl network list\|get\|create\|update\|delete` | Manage networks |
| `uictl wifi list\|get\|create\|update\|delete` | Manage WiFi broadcasts |
| `uictl hotspot list\|get\|create\|delete` | Manage guest vouchers |
| `uictl firewall zone list\|get\|create\|update\|delete` | Manage firewall zones |
| `uictl firewall policy list\|get\|create\|update\|delete\|order` | Manage firewall policies |
| `uictl acl list\|get\|create\|update\|delete\|order` | Manage ACL rules |
| `uictl dns list\|get\|create\|update\|delete` | Manage DNS policies |
| `uictl traffic-list list\|get\|create\|update\|delete` | Manage traffic matching lists |
| `uictl switching lag\|stack\|mc-lag list\|get` | View switching configs |
| `uictl country\|dpi\|device-tag\|radius\|wan\|vpn list` | Supporting resources |
| `uictl camera list\|get\|update\|snapshot\|stream\|ptz` | Manage Protect cameras |
| `uictl light list\|get\|update` | Manage Protect lights |
| `uictl sensor list\|get\|update` | Manage Protect sensors |
| `uictl chime list\|get\|update` | Manage Protect chimes |
| `uictl viewer list\|get\|update` | Manage Protect viewers |
| `uictl liveview list\|get\|create\|update` | Manage Protect live views |
| `uictl nvr get` | View NVR details |
| `uictl alarm webhook <trigger-id>` | Trigger Protect alarms |
| `uictl protect-info` | Protect application info |
| `uictl api <method> <path>` | Raw API passthrough |
| `uictl config show\|set\|path` | Manage configuration |
| `uictl schema [resource.action]` | Runtime schema introspection |
| `uictl skills` | Agent-optimized usage instructions |
| `uictl version` | Version information |

## Output Formats

| Context | Default | Override |
|---|---|---|
| Interactive terminal (TTY) | Colored table | `--json`, `--csv`, `--output ndjson` |
| Piped / non-TTY / agent | JSON | `--output table` to force table |
| Environment variable | — | `UICTL_OUTPUT_FORMAT=json` |

Use `--fields id,name,status` to select specific fields (works with all formats).

## LLM/Agent Integration

uictl is designed to work seamlessly with AI coding agents:

```bash
# Agents should start here — discover available commands
uictl schema
uictl schema network.create

# Or read the full skills reference
uictl skills

# Field masks minimize token usage
uictl device list --fields id,name,state

# Dry-run before mutations
uictl network delete <id> --dry-run
uictl network delete <id> --yes
```

Errors include structured guidance on stderr:
```json
{
  "code": "AUTH_EXPIRED",
  "message": "Session token has expired",
  "guidance": "Run `uictl login` to re-authenticate, then retry the command."
}
```

See [SKILLS.md](SKILLS.md) for the full agent reference.

## Configuration

```bash
# Interactive setup
uictl login

# Or set via environment variables
export UICTL_HOST=192.168.1.1
export UICTL_API_KEY=your-api-key
export UICTL_SITE=default

# Config file: ~/.config/uictl/config.yaml
uictl config show
uictl config set host 192.168.1.1

# Named profiles for multiple controllers
uictl login --profile office
uictl device list --profile office
```

## API Reference

See [docs/api-reference.md](docs/api-reference.md) for the full UniFi API reference (Network v10.2.93, Protect v7.0.88).

## License

MIT
