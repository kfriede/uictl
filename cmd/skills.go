package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(skillsCmd)
}

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Dump agent-optimized usage instructions",
	Long: `Prints concise, agent-optimized usage instructions for LLM agents.

This is designed to be read once and internalized by an agent to
reduce hallucinations and improve command accuracy.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(cmd.OutOrStdout(), skillsText)
		return nil
	},
}

const skillsText = `# uictl — Agent Skills

## Quick Reference
uictl <resource> <action> [flags]
Resources: site, device, client, network, wifi, hotspot, firewall, acl, dns, traffic-list, switching, country, dpi, device-tag, radius, wan, vpn

## Rules
- ALWAYS use --fields on list/get to limit output (saves tokens)
- ALWAYS use --dry-run before any mutating command, then confirm with user
- ALWAYS pass --yes on confirmed destructive actions (delete, remove)
- Use --json-input for complex create/update payloads (avoids flag hallucination)
- Parse JSON from stdout; errors go to stderr as JSON with "guidance" field
- Non-TTY automatically outputs JSON — no need for --json in agent context

## Invariants
- Resource names are singular nouns: device, client, network, site
- All timestamps are ISO 8601 / UTC
- MAC addresses are lowercase colon-separated: aa:bb:cc:dd:ee:ff
- IDs are UUIDs unless otherwise noted
- Pagination is handled automatically (all results returned by default)

## Common Patterns

### List with field selection
uictl device list --fields id,name,model,state
uictl client list --fields id,name,ipAddress,type
uictl network list --fields id,name,vlanId,enabled

### Create with JSON input (preferred for agents)
uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'
uictl hotspot create --json-input '{"name":"Day Pass","timeLimitMinutes":1440,"count":10}'

### Mutating with dry-run first
uictl network create --dry-run --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'
# show preview to user, then:
uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'

### Destructive with confirmation
uictl network delete <network-id> --dry-run
# show preview to user, then:
uictl network delete <network-id> --yes

### Raw API passthrough (for unimplemented endpoints)
uictl api get /v1/info
uictl api post /v1/sites/{siteId}/devices --data '{"macAddress":"aa:bb:cc:dd:ee:ff"}'

### Runtime introspection
uictl schema                      # list all commands
uictl schema network.create       # full schema for a specific command

## Error Handling
Errors include structured JSON on stderr:
{"code":"AUTH_EXPIRED","message":"Session token has expired","guidance":"Run 'uictl login' to re-authenticate."}

Exit codes: 0=success, 1=general error, 2=auth error, 3=not found, 4=conflict/validation

## Configuration
UICTL_HOST, UICTL_SITE, UICTL_API_KEY, UICTL_OUTPUT_FORMAT=json, UICTL_DEBUG=1
`
