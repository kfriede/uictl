<p align="center">
  <strong>uictl</strong> — manage your UniFi network from the terminal
</p>

<p align="center">
  <a href="#install">Install</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#commands">Commands</a> •
  <a href="#for-llm-agents">For LLM Agents</a> •
  <a href="#why-uictl">Why uictl?</a>
</p>

---

**uictl** is a CLI for Ubiquiti UniFi controllers. It covers both the Network and Protect APIs — 108 operations total — in a single static binary. Designed from the ground up for humans *and* LLM agents.

```bash
# See your network at a glance
$ uictl device list --fields name,model,state
NAME              MODEL               STATE
────────────────  ──────────────────  ──────
wap01-attic       U6 Pro              ONLINE
sw01-attic        USW Pro Max 16 PoE  ONLINE
Home UDMP         UDM Pro             ONLINE

# Pipe to an agent? Automatic JSON — no flags needed
$ uictl device list --fields name,state | jq '.[].name'
"wap01-attic"
"sw01-attic"
"Home UDMP"
```

## Why uictl?

There are other UniFi CLIs. Here's why uictl is different:

| Feature | uictl | unified | unifly | ui-cli |
|---|:---:|:---:|:---:|:---:|
| Network API (full) | ✅ | ✅ | ✅ | ✅ |
| Protect API (full) | ✅ | ✅ | ❌ | ❌ |
| Auto-JSON for agents (non-TTY) | ✅ | ❌ | ❌ | ❌ |
| `--fields` (token-efficient output) | ✅ | ❌ | ❌ | ❌ |
| `schema` command (runtime introspection) | ✅ | ❌ | ❌ | ❌ |
| Structured error guidance | ✅ | ❌ | ❌ | ❌ |
| `--dry-run` on every mutation | ✅ | ❌ | ❌ | ❌ |
| `--json-input` (no flag hallucination) | ✅ | ❌ | ❌ | ❌ |
| Agent skills file (SKILLS.md) | ✅ | ⚠️ | ❌ | ❌ |
| NDJSON streaming output | ✅ | ❌ | ❌ | ❌ |
| UniFi OS auto-detection | ✅ | ❌ | ✅ | ✅ |
| OS keyring credentials | ✅ | ❌ | ✅ | ❌ |
| Single static binary | ✅ | ✅ | ✅ | ❌ |
| Language | Go | Go | Rust | Python |

**The core differentiator**: uictl treats LLM agents as first-class users. Every design decision — output format, error messages, input handling, safety rails — considers both the human at the keyboard and the agent in the pipe.

## Install

**Homebrew** (macOS and Linux):

```bash
brew install kfriede/tap/uictl
```

**Pre-built binaries** (linux, macOS, Windows — amd64 + arm64):

Download from [GitHub Releases](https://github.com/kfriede/uictl/releases/latest).

**Go install**:

```bash
go install github.com/kfriede/uictl@latest
```

**From source**:

```bash
git clone https://github.com/kfriede/uictl.git && cd uictl && make build
```

## Quick Start

```bash
# 1. Authenticate (interactive — stores API key in OS keyring)
uictl login

# 2. List your sites
uictl site list

# 3. See your devices
uictl device list --fields name,model,state,ipAddress

# 4. Check a device's health
uictl device stats <device-id>

# 5. Create a guest voucher
uictl hotspot create --name "Day Pass" --duration 1440 --count 10

# 6. Manage networks
uictl network create --name "IoT" --vlan 30
uictl network update <id> --json-input '{"name":"IoT v2","enabled":true,"management":false,"vlanId":30}'
uictl network delete <id> --yes

# 7. Take a camera snapshot
uictl camera snapshot <camera-id> > front-door.jpg

# 8. Anything the CLI doesn't cover yet — raw API passthrough
uictl api get /v1/info
```

## Commands

### Network API (73 operations)

| Resource | Actions |
|---|---|
| `site` | `list` |
| `device` | `list` `get` `adopt` `remove` `restart` `stats` `pending` `port-action` |
| `client` | `list` `get` `authorize` `unauthorize` |
| `network` | `list` `get` `create` `update` `delete` `references` |
| `wifi` | `list` `get` `create` `update` `delete` |
| `hotspot` | `list` `get` `create` `delete` |
| `firewall zone` | `list` `get` `create` `update` `delete` |
| `firewall policy` | `list` `get` `create` `update` `delete` `order` |
| `acl` | `list` `get` `create` `update` `delete` `order` |
| `dns` | `list` `get` `create` `update` `delete` |
| `traffic-list` | `list` `get` `create` `update` `delete` |
| `switching` | `lag list\|get` `stack list\|get` `mc-lag list\|get` |
| `country` `dpi` `device-tag` `radius` `wan` `vpn` | `list` |

### Protect API (35 operations)

| Resource | Actions |
|---|---|
| `camera` | `list` `get` `update` `snapshot` `stream` `stream-create` `stream-delete` `talkback` `disable-mic` `ptz goto\|patrol-start\|patrol-stop` |
| `light` | `list` `get` `update` |
| `sensor` | `list` `get` `update` |
| `chime` | `list` `get` `update` |
| `viewer` | `list` `get` `update` |
| `liveview` | `list` `get` `create` `update` |
| `nvr` | `get` |
| `alarm` | `webhook` |

### Utility

| Command | Description |
|---|---|
| `login` | Interactive auth (stores key in OS keyring) |
| `config show\|set\|path` | View and manage configuration |
| `api <method> <path>` | Raw API passthrough |
| `schema [resource.action]` | Runtime command introspection (for agents) |
| `skills` | Agent-optimized usage instructions |
| `version` | Version info (supports `--json`) |
| `completion bash\|zsh\|fish` | Shell completions with dynamic API lookups |

## Output

uictl auto-detects what you need:

| Context | What you get |
|---|---|
| **Terminal (TTY)** | Colored, aligned tables |
| **Piped / scripted / agent** | JSON (automatic, no flags needed) |
| `--json` | Force JSON anywhere |
| `--csv` | CSV output |
| `--output ndjson` | One JSON object per line (streaming) |
| `--fields name,ip,state` | Only the fields you ask for |
| `UICTL_OUTPUT_FORMAT=json` | Set globally via env var |

**stdout** is always data. Logs, progress, and errors go to **stderr**.

## Safety

Every mutating command supports `--dry-run`:

```bash
$ uictl network delete <id> --dry-run
[dry-run] Would delete network a69e9a69-8bd0-49b4-8f65-42345bf8e8ec

$ uictl network delete <id>
Are you sure you want to delete network a69e9a69? (y/N):

$ uictl network delete <id> --yes    # skip prompt (for scripts/agents)
✓ Deleted network a69e9a69-8bd0-49b4-8f65-42345bf8e8ec
```

## For LLM Agents

uictl is built to be the CLI that agents don't fight with. Here's why:

### No config needed — just env vars

```bash
export UICTL_HOST=192.168.1.1
export UICTL_API_KEY=your-key
export UICTL_SITE=default
# That's it. Every command works now.
```

### Auto-JSON when piped

Agents never see table output. When stdout isn't a TTY, uictl automatically outputs JSON:

```bash
# Agent runs this — gets JSON, not a table
uictl device list --fields id,name,state
```

### Schema introspection

Agents discover commands at runtime instead of hallucinating flags:

```bash
$ uictl schema network.create
{
  "resource": "network",
  "action": "create",
  "httpMethod": "POST",
  "apiPath": "/v1/sites/{siteId}/networks",
  "flags": [
    {"name": "name", "type": "string", "description": "Network name"},
    {"name": "vlan", "type": "integer", "description": "VLAN ID"},
    {"name": "json-input", "type": "string", "description": "Full JSON request body (preferred for agents)"}
  ],
  "example": "uictl network create --json-input '{\"name\":\"IoT\",\"enabled\":true,\"management\":false,\"vlanId\":30}'",
  "mutating": true,
  "supportsDryRun": true
}
```

### `--json-input` prevents flag hallucination

Instead of guessing flags, agents send the exact API payload:

```bash
uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'
```

### `--fields` minimizes token usage

```bash
# 5 fields instead of 40 — saves tokens, faster parsing
uictl device list --fields id,name,model,state,ipAddress
```

### Structured errors with guidance

When something fails, the error tells the agent exactly what to do next:

```json
{
  "code": "AUTH_EXPIRED",
  "message": "Session token has expired",
  "guidance": "Run `uictl login` to re-authenticate, then retry the command."
}
```

### Safe by default

- `--dry-run` on **every** mutation — agents preview before executing
- `--yes` required for destructive actions in non-TTY (never hangs waiting for input)
- Input validation rejects malformed IDs with clear error messages

### Agent config files

uictl ships config files that agents discover automatically:

| File | Agent | Purpose |
|---|---|---|
| [`AGENTS.md`](AGENTS.md) | All agents (cross-vendor standard) | Full usage spec, rules, patterns, boundaries |
| [`CLAUDE.md`](CLAUDE.md) | Claude Code / Claude Desktop | Points to AGENTS.md + quick reference |
| [`.github/copilot-instructions.md`](.github/copilot-instructions.md) | GitHub Copilot CLI | Project conventions and design principles |
| [`.claude-plugin/`](.claude-plugin/) | Claude Cowork marketplace | Installable plugin with skills |
| [`SKILLS.md`](SKILLS.md) | Any agent (via `uictl skills`) | YAML frontmatter + usage patterns |

Agents that clone or work within this repo will automatically pick up the appropriate file. For agents using uictl as an *external tool* (not within the repo), set the env vars and run `uictl skills` to bootstrap.

### Claude Cowork Plugin

Install uictl as a Claude Cowork plugin for natural language network management:

```
/plugin install https://github.com/kfriede/uictl
```

Then ask Claude things like:
- "List all my UniFi devices"
- "Create a guest WiFi network with a 1-day voucher"
- "Show me the firewall rules between my IoT and Server zones"
- "Take a snapshot from the front door camera"

## Configuration

```bash
# Interactive setup (stores API key in OS keyring)
uictl login

# Or configure via environment
export UICTL_HOST=192.168.1.1
export UICTL_API_KEY=your-api-key
export UICTL_SITE=default

# Or config file (~/.config/uictl/config.yaml)
uictl config set host 192.168.1.1
uictl config show

# Multiple controllers with named profiles
uictl login --profile office
uictl login --profile home
uictl device list --profile office
```

**Precedence**: CLI flags > environment variables > config file.

### UniFi OS Auto-Detection

uictl automatically detects whether your controller is a UniFi OS console (UDM, UDM Pro, etc.) or a standalone Network Application and uses the correct API path. No configuration needed.

### Smart Site Resolution

Use site names, not just UUIDs:

```bash
uictl device list --site default     # resolves to UUID automatically
uictl device list --site "My Site"   # works too
```

## API Reference

See [docs/api-reference.md](docs/api-reference.md) for the full UniFi API reference (Network v10.2.93, Protect v7.0.88).

## Contributing

```bash
git clone https://github.com/kfriede/uictl.git
cd uictl
make all        # lint + test + build
make test       # just tests
make lint       # just lint
```

## License

MIT
