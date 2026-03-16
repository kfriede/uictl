---
tool: uictl
version: 0.1.0
description: CLI for managing UniFi network devices and controllers
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
  - IDs are UUIDs unless otherwise noted
  - non-TTY mode automatically outputs JSON
---

# uictl Agent Skills

## Command Pattern
```
uictl <resource> <action> [flags]
```

## Resources
| Resource | Actions | Description |
|---|---|---|
| site | list | UniFi sites |
| device | list, get, adopt, remove, restart, stats, pending, port-action | Network devices |
| client | list, get, authorize, unauthorize | Connected clients |
| network | list, get, create, update, delete, references | VLANs and subnets |
| wifi | list, get, create, update, delete | WiFi broadcasts (SSIDs) |
| hotspot | list, get, create, delete | Guest vouchers |
| firewall zone | list, get, create, update, delete | Firewall zones |
| firewall policy | list, get, create, update, delete, order | Firewall policies |
| acl | list, get, create, update, delete, order | ACL rules |
| dns | list, get, create, update, delete | DNS policies |
| traffic-list | list, get, create, update, delete | Traffic matching lists |
| switching lag | list, get | LAG configs (read-only) |
| switching stack | list, get | Switch stacks (read-only) |
| switching mc-lag | list, get | MC-LAG domains (read-only) |
| country | list | Country codes |
| dpi | app, category | DPI reference data |
| device-tag | list | Device tags |
| radius | list | RADIUS profiles |
| wan | list | WAN interfaces |
| vpn | server, tunnel | VPN servers and tunnels |

## Global Flags
| Flag | Short | Description |
|---|---|---|
| --json | | Force JSON output |
| --csv | | Force CSV output |
| --output | -o | table, json, csv, ndjson |
| --fields | | Comma-separated field mask |
| --quiet | -q | Suppress non-essential output |
| --site | -s | Site name or ID |
| --profile | -p | Configuration profile |
| --yes | -y | Skip confirmation prompts |
| --dry-run | | Preview without executing |
| --insecure | -k | Skip TLS verification |
| --verbose | -v | Verbose logging to stderr |
| --debug | | Full request/response logging |
| --no-color | | Disable colors |

## Common Workflows

### Discover available commands
```bash
uictl schema
uictl schema device.list
```

### List devices with minimal fields
```bash
uictl device list --fields id,name,model,state
```

### Create a network (agent-preferred pattern)
```bash
# Step 1: dry-run
uictl network create --dry-run --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'
# Step 2: execute after user confirms
uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'
```

### Delete with safety
```bash
# Step 1: preview
uictl network delete <id> --dry-run
# Step 2: execute
uictl network delete <id> --yes
```

### Generate guest vouchers
```bash
uictl hotspot create --name "Event Pass" --duration 480 --count 20
```

### Raw API access
```bash
uictl api get /v1/info
uictl api post /v1/sites/{siteId}/devices --data '{"macAddress":"aa:bb:cc:dd:ee:ff","ignoreDeviceLimit":false}'
```

## Error Format (stderr)
```json
{
  "code": "AUTH_EXPIRED",
  "message": "Session token has expired",
  "guidance": "Run `uictl login` to re-authenticate, then retry the command."
}
```

## Exit Codes
| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error |
| 2 | Authentication/permission error |
| 3 | Resource not found |
| 4 | Conflict/validation error |
