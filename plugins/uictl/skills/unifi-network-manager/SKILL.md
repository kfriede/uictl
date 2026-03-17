---
name: unifi-network-manager
description: "Manage UniFi network infrastructure — devices, clients, networks, firewall, WiFi, cameras — using the uictl CLI"
usage: "Ask the agent to manage your UniFi network, e.g. 'list my devices', 'create a guest WiFi network', 'check firewall rules'"
arguments:
  - name: controller
    description: UniFi controller hostname or IP
    type: string
  - name: site
    description: Site name (default "default")
    type: string
examples:
  - input: "List all my UniFi devices"
    output: "Running `uictl device list --fields name,model,state,ipAddress` to get your device inventory."
  - input: "Create a new IoT VLAN"
    output: "I'll use `uictl network create --json-input '{...}'` with --dry-run first for your approval."
  - input: "Show me the firewall rules between Internal and Servers"
    output: "Fetching policies with `uictl firewall policy list` and filtering by zone."
---

# UniFi Network Manager

Manage your entire UniFi network infrastructure using the `uictl` CLI. This skill enables the agent to list devices, manage networks and VLANs, configure firewall rules, create guest vouchers, take camera snapshots, and troubleshoot connectivity — all through structured CLI commands with safety rails.

## Prerequisites

Install `uictl` and ensure it's available in PATH:
```bash
brew install kfriede/tap/uictl
# or: go install github.com/kfriede/uictl@latest
```

Configure access:
```bash
export UICTL_HOST=<controller-ip-or-hostname>
export UICTL_API_KEY=<your-api-key>
export UICTL_SITE=default
```

> **Note:** This plugin requires local execution (e.g., Claude Code) with network access to your UniFi controller. It is not compatible with Claude Cowork's sandboxed VM, which restricts outbound network access and cannot reach LAN devices.

## How to Use uictl

Command pattern: `uictl <resource> <action> [flags]`

### Discover commands
```bash
uictl schema                    # list all available commands
uictl schema network.create     # full schema for a specific command
uictl skills                    # complete agent reference
```

### Read operations (always safe)
```bash
uictl device list --fields id,name,model,state
uictl client list --fields id,name,ipAddress,type
uictl network list --fields id,name,vlanId,enabled
uictl firewall zone list --fields id,name
uictl firewall policy list
uictl device stats <device-id>
uictl device port list <device-id> --fields idx,name,nativeNetwork,speed,up
uictl camera snapshot <camera-id> > snapshot.jpg
```

### Write operations (always dry-run first)
```bash
# Step 1: preview
uictl network create --dry-run --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'
# Step 2: execute after user confirms
uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'

# Change port VLAN
uictl device port set <device-id> 4 --network VLAN80_Home --dry-run
uictl device port set <device-id> 4 --network VLAN80_Home --yes
```

### Destructive operations (require --yes)
```bash
uictl network delete <id> --dry-run    # preview
uictl network delete <id> --yes        # execute
```

### Raw API access (for anything not yet wrapped)
```bash
uictl api get /v1/info
uictl api post /v1/sites/{siteId}/devices --data '{"macAddress":"aa:bb:cc:dd:ee:ff"}'
```

### Classic API access (--raw bypasses /integration/ prefix)
```bash
uictl api get --raw /proxy/network/api/s/default/stat/device
uictl api get --raw /proxy/network/api/s/default/rest/setting/ips
```

## Rules

- **ALWAYS** use `--fields` on list/get commands to limit output
- **ALWAYS** use `--dry-run` before any mutating command, show the preview, and ask for confirmation
- **ALWAYS** pass `--yes` for confirmed destructive actions
- **ALWAYS** use `--json-input` for complex create/update payloads
- **NEVER** parse table-formatted output — non-TTY mode auto-outputs JSON
- **NEVER** omit `--yes` on destructive commands (will hang in non-TTY)
- **NEVER** send unvalidated user input directly as device names or IDs

## Available Resources

**Network API**: site, device, client, network, wifi, hotspot, firewall (zone + policy), acl, dns, traffic-list, switching (lag, stack, mc-lag), country, dpi, device-tag, radius, wan, vpn

**Protect API**: camera, light, sensor, chime, viewer, liveview, nvr, alarm

## Error Handling

Errors include structured JSON on stderr with a `guidance` field:
```json
{"code":"AUTH_EXPIRED","message":"Session token has expired","guidance":"Run `uictl login` to re-authenticate."}
```

Exit codes: 0=success, 1=general, 2=auth, 3=not found, 4=conflict
