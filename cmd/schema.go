package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(schemaCmd)
}

// SchemaEntry describes a single command for agent introspection.
type SchemaEntry struct {
	Resource    string        `json:"resource"`
	Action      string        `json:"action"`
	Description string        `json:"description"`
	Method      string        `json:"httpMethod"`
	Path        string        `json:"apiPath"`
	Parameters  []SchemaParam `json:"parameters,omitempty"`
	Flags       []SchemaFlag  `json:"flags,omitempty"`
	Example     string        `json:"example"`
	Mutating    bool          `json:"mutating"`
	DryRun      bool          `json:"supportsDryRun"`
}

// SchemaParam describes a path/query parameter.
type SchemaParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	In       string `json:"in"` // path, query
}

// SchemaFlag describes a CLI flag.
type SchemaFlag struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Desc     string `json:"description"`
}

var schemaCmd = &cobra.Command{
	Use:   "schema [resource.action]",
	Short: "Runtime command schema for LLM agents",
	Long: `Returns a JSON schema for any command, including parameters, types,
required fields, and copy-pasteable examples.

This is the primary entry point for LLM agents discovering how to use uictl.

Examples:
  uictl schema                       List all available commands
  uictl schema device.list           Schema for device list
  uictl schema network.create        Schema for network create
  uictl schema firewall.policy.list  Schema for firewall policy list`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registry := buildSchemaRegistry()

		if len(args) == 0 {
			// List all commands
			summary := make([]map[string]string, 0, len(registry))
			for key, entry := range registry {
				summary = append(summary, map[string]string{
					"command":     key,
					"description": entry.Description,
					"mutating":    fmt.Sprintf("%v", entry.Mutating),
				})
			}
			return printAPIResult(toAnySlice(mapSlice(summary)))
		}

		key := args[0]
		entry, ok := registry[key]
		if !ok {
			// Try to suggest
			suggestions := findSimilar(key, registry)
			msg := fmt.Sprintf("unknown command: %s", key)
			if len(suggestions) > 0 {
				msg += fmt.Sprintf("\n\nDid you mean one of these?\n  %s", strings.Join(suggestions, "\n  "))
			}
			return fmt.Errorf("%s", msg)
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(entry)
	},
}

func buildSchemaRegistry() map[string]SchemaEntry {
	siteParam := SchemaParam{Name: "siteId", Type: "string", Required: true, In: "config"}

	r := map[string]SchemaEntry{
		"info": {
			Resource: "info", Action: "get", Description: "Get controller application info",
			Method: "GET", Path: "/v1/info", Example: "uictl info",
		},
		"site.list": {
			Resource: "site", Action: "list", Description: "List local sites",
			Method: "GET", Path: "/v1/sites", Example: "uictl site list",
			Parameters: []SchemaParam{{Name: "offset", Type: "integer", In: "query"}, {Name: "limit", Type: "integer", In: "query"}},
		},
		"device.list": {
			Resource: "device", Action: "list", Description: "List adopted devices",
			Method: "GET", Path: "/v1/sites/{siteId}/devices", Example: "uictl device list --fields id,name,model,state",
			Parameters: []SchemaParam{siteParam},
		},
		"device.get": {
			Resource: "device", Action: "get", Description: "Get adopted device details",
			Method: "GET", Path: "/v1/sites/{siteId}/devices/{deviceId}", Example: "uictl device get <device-id>",
			Parameters: []SchemaParam{siteParam, {Name: "deviceId", Type: "string", Required: true, In: "path"}},
		},
		"device.adopt": {
			Resource: "device", Action: "adopt", Description: "Adopt a pending device",
			Method: "POST", Path: "/v1/sites/{siteId}/devices", Mutating: true, DryRun: true,
			Example: "uictl device adopt --mac aa:bb:cc:dd:ee:ff",
			Flags: []SchemaFlag{
				{Name: "mac", Type: "string", Desc: "MAC address of device to adopt"},
				{Name: "json-input", Type: "string", Desc: "Full JSON request body"},
			},
		},
		"device.remove": {
			Resource: "device", Action: "remove", Description: "Remove (unadopt) a device",
			Method: "DELETE", Path: "/v1/sites/{siteId}/devices/{deviceId}", Mutating: true, DryRun: true,
			Example: "uictl device remove <device-id> --yes",
		},
		"device.restart": {
			Resource: "device", Action: "restart", Description: "Restart a device",
			Method: "POST", Path: "/v1/sites/{siteId}/devices/{deviceId}/actions", Mutating: true, DryRun: true,
			Example: "uictl device restart <device-id>",
		},
		"device.stats": {
			Resource: "device", Action: "stats", Description: "Get latest device statistics",
			Method: "GET", Path: "/v1/sites/{siteId}/devices/{deviceId}/statistics/latest",
			Example: "uictl device stats <device-id>",
		},
		"device.pending": {
			Resource: "device", Action: "pending", Description: "List devices pending adoption",
			Method: "GET", Path: "/v1/pending-devices", Example: "uictl device pending",
		},
		"device.port-action": {
			Resource: "device", Action: "port-action", Description: "Execute a port action (e.g. cycle PoE)",
			Method: "POST", Path: "/v1/sites/{siteId}/devices/{deviceId}/interfaces/ports/{portIdx}/actions",
			Mutating: true, DryRun: true,
			Example: "uictl device port-action <device-id> 4 cycle",
			Parameters: []SchemaParam{
				siteParam,
				{Name: "deviceId", Type: "string", Required: true, In: "path"},
				{Name: "portIdx", Type: "integer", Required: true, In: "path"},
				{Name: "action", Type: "string", Required: true, In: "body"},
			},
		},
		"device.port.list": {
			Resource: "device", Action: "port.list", Description: "List switch ports with VLAN/network assignments",
			Method: "GET", Path: "/proxy/network/api/s/{site}/stat/device/{mac} (classic API)",
			Example: "uictl device port list <device-id> --fields idx,name,nativeNetwork,speed,poe,up",
			Parameters: []SchemaParam{
				siteParam,
				{Name: "deviceId", Type: "string", Required: true, In: "path"},
			},
		},
		"device.port.set": {
			Resource: "device", Action: "port.set", Description: "Set the native network (VLAN) on a switch port",
			Method: "PUT", Path: "/proxy/network/api/s/{site}/rest/device/{_id} (classic API)",
			Mutating: true, DryRun: true,
			Example: "uictl device port set <device-id> 4 --network VLAN80_Home",
			Parameters: []SchemaParam{
				siteParam,
				{Name: "deviceId", Type: "string", Required: true, In: "path"},
				{Name: "portIdx", Type: "integer", Required: true, In: "path"},
			},
			Flags: []SchemaFlag{
				{Name: "network", Type: "string", Desc: "Network name, VLAN ID, or classic network ID"},
				{Name: "json-input", Type: "string", Desc: "Full JSON port override object"},
			},
		},
		"device.port.action": {
			Resource: "device", Action: "port.action", Description: "Execute a port action (alias for port-action)",
			Method: "POST", Path: "/v1/sites/{siteId}/devices/{deviceId}/interfaces/ports/{portIdx}/actions",
			Mutating: true, DryRun: true,
			Example: "uictl device port action <device-id> 4 cycle",
		},
		"api": {
			Resource: "api", Action: "passthrough", Description: "Raw API request (use --raw for classic API endpoints)",
			Method: "any", Path: "user-specified",
			Example: "uictl api get /v1/info\nuictl api get --raw /proxy/network/api/s/default/stat/device",
			Flags: []SchemaFlag{
				{Name: "data", Type: "string", Desc: "JSON request body (-d)"},
				{Name: "raw", Type: "boolean", Desc: "Bypass /integration/ prefix for classic API"},
			},
		},
		"client.list": {
			Resource: "client", Action: "list", Description: "List connected clients",
			Method: "GET", Path: "/v1/sites/{siteId}/clients",
			Example:    "uictl client list --fields id,name,ipAddress,type",
			Parameters: []SchemaParam{siteParam},
		},
		"client.get": {
			Resource: "client", Action: "get", Description: "Get connected client details",
			Method: "GET", Path: "/v1/sites/{siteId}/clients/{clientId}",
			Example: "uictl client get <client-id>",
		},
		"client.authorize": {
			Resource: "client", Action: "authorize", Description: "Authorize a guest client",
			Method: "POST", Path: "/v1/sites/{siteId}/clients/{clientId}/actions", Mutating: true, DryRun: true,
			Example: "uictl client authorize <client-id>",
		},
		"client.unauthorize": {
			Resource: "client", Action: "unauthorize", Description: "Unauthorize a guest client",
			Method: "POST", Path: "/v1/sites/{siteId}/clients/{clientId}/actions", Mutating: true, DryRun: true,
			Example: "uictl client unauthorize <client-id>",
		},
		"network.list": {
			Resource: "network", Action: "list", Description: "List networks",
			Method: "GET", Path: "/v1/sites/{siteId}/networks",
			Example: "uictl network list --fields id,name,vlanId,enabled",
		},
		"network.get": {
			Resource: "network", Action: "get", Description: "Get network details",
			Method: "GET", Path: "/v1/sites/{siteId}/networks/{networkId}",
			Example: "uictl network get <network-id>",
		},
		"network.create": {
			Resource: "network", Action: "create", Description: "Create a network",
			Method: "POST", Path: "/v1/sites/{siteId}/networks", Mutating: true, DryRun: true,
			Example: `uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'`,
			Flags: []SchemaFlag{
				{Name: "name", Type: "string", Desc: "Network name"},
				{Name: "vlan", Type: "integer", Desc: "VLAN ID"},
				{Name: "json-input", Type: "string", Desc: "Full JSON request body (preferred for agents)"},
			},
		},
		"network.update": {
			Resource: "network", Action: "update", Description: "Update a network",
			Method: "PUT", Path: "/v1/sites/{siteId}/networks/{networkId}", Mutating: true, DryRun: true,
			Example: `uictl network update <id> --json-input '{"name":"IoT v2","enabled":true,"management":false,"vlanId":30}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "Full JSON request body"}},
		},
		"network.delete": {
			Resource: "network", Action: "delete", Description: "Delete a network",
			Method: "DELETE", Path: "/v1/sites/{siteId}/networks/{networkId}", Mutating: true, DryRun: true,
			Example: "uictl network delete <network-id> --yes",
			Flags:   []SchemaFlag{{Name: "force", Type: "boolean", Desc: "Force delete even if in use"}},
		},
		"wifi.list": {
			Resource: "wifi", Action: "list", Description: "List WiFi broadcasts",
			Method: "GET", Path: "/v1/sites/{siteId}/wifi/broadcasts", Example: "uictl wifi list",
		},
		"wifi.create": {
			Resource: "wifi", Action: "create", Description: "Create a WiFi broadcast",
			Method: "POST", Path: "/v1/sites/{siteId}/wifi/broadcasts", Mutating: true, DryRun: true,
			Example: `uictl wifi create --json-input '{"name":"Guest WiFi",...}'`,
		},
		"wifi.delete": {
			Resource: "wifi", Action: "delete", Description: "Delete a WiFi broadcast",
			Method: "DELETE", Path: "/v1/sites/{siteId}/wifi/broadcasts/{id}", Mutating: true, DryRun: true,
			Example: "uictl wifi delete <broadcast-id> --yes",
		},
		"hotspot.list": {
			Resource: "hotspot", Action: "list", Description: "List hotspot vouchers",
			Method: "GET", Path: "/v1/sites/{siteId}/hotspot/vouchers", Example: "uictl hotspot list",
		},
		"hotspot.create": {
			Resource: "hotspot", Action: "create", Description: "Generate hotspot vouchers",
			Method: "POST", Path: "/v1/sites/{siteId}/hotspot/vouchers", Mutating: true, DryRun: true,
			Example: `uictl hotspot create --name "Day Pass" --duration 1440 --count 10`,
			Flags: []SchemaFlag{
				{Name: "name", Type: "string", Required: true, Desc: "Voucher name"},
				{Name: "duration", Type: "integer", Required: true, Desc: "Duration in minutes"},
				{Name: "count", Type: "integer", Desc: "Number of vouchers"},
				{Name: "json-input", Type: "string", Desc: "Full JSON request body"},
			},
		},
		"firewall.zone.list": {
			Resource: "firewall", Action: "zone.list", Description: "List firewall zones",
			Method: "GET", Path: "/v1/sites/{siteId}/firewall/zones", Example: "uictl firewall zone list",
		},
		"firewall.zone.create": {
			Resource: "firewall", Action: "zone.create", Description: "Create a firewall zone",
			Method: "POST", Path: "/v1/sites/{siteId}/firewall/zones", Mutating: true, DryRun: true,
			Example: `uictl firewall zone create --json-input '{"name":"IoT","networkIds":["<id>"]}'`,
		},
		"firewall.policy.list": {
			Resource: "firewall", Action: "policy.list", Description: "List firewall policies",
			Method: "GET", Path: "/v1/sites/{siteId}/firewall/policies", Example: "uictl firewall policy list",
		},
		"firewall.policy.create": {
			Resource: "firewall", Action: "policy.create", Description: "Create a firewall policy",
			Method: "POST", Path: "/v1/sites/{siteId}/firewall/policies", Mutating: true, DryRun: true,
			Example: `uictl firewall policy create --json-input '{"name":"Block IoT","enabled":true,"action":{"type":"BLOCK"},...}'`,
		},
		"acl.list": {
			Resource: "acl", Action: "list", Description: "List ACL rules",
			Method: "GET", Path: "/v1/sites/{siteId}/acl-rules", Example: "uictl acl list",
		},
		"dns.list": {
			Resource: "dns", Action: "list", Description: "List DNS policies",
			Method: "GET", Path: "/v1/sites/{siteId}/dns/policies", Example: "uictl dns list",
		},
		"traffic-list.list": {
			Resource: "traffic-list", Action: "list", Description: "List traffic matching lists",
			Method: "GET", Path: "/v1/sites/{siteId}/traffic-matching-lists", Example: "uictl traffic-list list",
		},
	}

	return r
}

func findSimilar(input string, registry map[string]SchemaEntry) []string {
	var suggestions []string
	inputLower := strings.ToLower(input)
	for key := range registry {
		if strings.Contains(key, inputLower) || strings.Contains(inputLower, strings.Split(key, ".")[0]) {
			suggestions = append(suggestions, key)
		}
	}
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	return suggestions
}

func mapSlice(maps []map[string]string) []map[string]any {
	result := make([]map[string]any, len(maps))
	for i, m := range maps {
		r := make(map[string]any, len(m))
		for k, v := range m {
			r[k] = v
		}
		result[i] = r
	}
	return result
}
