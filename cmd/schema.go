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
		"alarm-hub.list": {
			Resource: "alarm-hub", Action: "list", Description: "List Protect alarm hubs",
			Method: "GET", Path: "/v1/alarm-hubs", Example: "uictl alarm-hub list",
		},
		"alarm-hub.get": {
			Resource: "alarm-hub", Action: "get", Description: "Get Protect alarm hub details",
			Method: "GET", Path: "/v1/alarm-hubs/{id}", Example: "uictl alarm-hub get <id>",
		},
		"alarm-hub.update": {
			Resource: "alarm-hub", Action: "update", Description: "Update Protect alarm hub settings",
			Method: "PATCH", Path: "/v1/alarm-hubs/{id}", Mutating: true, DryRun: true,
			Example: `uictl alarm-hub update <id> --json-input '{"name":"Warehouse"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"alarm-hub.trigger": {
			Resource: "alarm-hub", Action: "trigger", Description: "Trigger a Protect alarm hub output",
			Method: "POST", Path: "/v1/alarm-hubs/{id}/outputs/{outputId}/trigger", Mutating: true, DryRun: true,
			Example: "uictl alarm-hub trigger <id> <output-id> --enable --duration 5000",
			Flags: []SchemaFlag{
				{Name: "enable", Type: "bool", Desc: "Set output on or off; omit to toggle"},
				{Name: "delay", Type: "integer", Desc: "Delay in milliseconds"},
				{Name: "duration", Type: "integer", Desc: "Duration in milliseconds"},
			},
		},
		"arm-profile.list": {
			Resource: "arm-profile", Action: "list", Description: "List Protect arm profiles",
			Method: "GET", Path: "/v1/arm-profiles", Example: "uictl arm-profile list",
		},
		"arm-profile.create": {
			Resource: "arm-profile", Action: "create", Description: "Create a Protect arm profile",
			Method: "POST", Path: "/v1/arm-profiles", Mutating: true, DryRun: true,
			Example: `uictl arm-profile create --json-input '{"name":"Away","automations":[],"schedules":[],"recordEverything":false,"activationDelay":0}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON body (name, automations, schedules, recordEverything, activationDelay)"}},
		},
		"arm-profile.update": {
			Resource: "arm-profile", Action: "update", Description: "Update a Protect arm profile",
			Method: "PATCH", Path: "/v1/arm-profiles/{id}", Mutating: true, DryRun: true,
			Example: `uictl arm-profile update <id> --json-input '{"name":"Away"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"arm-profile.delete": {
			Resource: "arm-profile", Action: "delete", Description: "Delete a Protect arm profile",
			Method: "DELETE", Path: "/v1/arm-profiles/{id}", Mutating: true, DryRun: true,
			Example: "uictl arm-profile delete <id> --yes",
		},
		"arm-profile.enable": {
			Resource: "arm-profile", Action: "enable", Description: "Enable the Protect arm alarm",
			Method: "POST", Path: "/v1/arm-profiles/enable", Mutating: true, DryRun: true,
			Example: "uictl arm-profile enable",
		},
		"arm-profile.disable": {
			Resource: "arm-profile", Action: "disable", Description: "Disable the Protect arm alarm",
			Method: "POST", Path: "/v1/arm-profiles/disable", Mutating: true, DryRun: true,
			Example: "uictl arm-profile disable",
		},
		"arm-profile.settings": {
			Resource: "arm-profile", Action: "settings", Description: "Set the current Protect arm profile",
			Method: "PATCH", Path: "/v1/arm-profiles/settings", Mutating: true, DryRun: true,
			Example: "uictl arm-profile settings <arm-profile-id>",
		},
		"bridge.list": {
			Resource: "bridge", Action: "list", Description: "List Protect bridges",
			Method: "GET", Path: "/v1/bridges", Example: "uictl bridge list",
		},
		"bridge.get": {
			Resource: "bridge", Action: "get", Description: "Get Protect bridge details",
			Method: "GET", Path: "/v1/bridges/{id}", Example: "uictl bridge get <id>",
		},
		"bridge.update": {
			Resource: "bridge", Action: "update", Description: "Update Protect bridge settings",
			Method: "PATCH", Path: "/v1/bridges/{id}", Mutating: true, DryRun: true,
			Example: `uictl bridge update <id> --json-input '{"name":"Garage"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"file.list": {
			Resource: "file", Action: "list", Description: "List Protect device asset files",
			Method: "GET", Path: "/v1/files/{fileType}", Example: "uictl file list <file-type>",
		},
		"file.upload": {
			Resource: "file", Action: "upload", Description: "Upload a Protect device asset file",
			Method: "POST", Path: "/v1/files/{fileType}", Mutating: true, DryRun: true,
			Example: "uictl file upload <file-type> --file ./tone.mp3",
			Flags:   []SchemaFlag{{Name: "file", Type: "string", Required: true, Desc: "Path of the file to upload"}},
		},
		"fob.list": {
			Resource: "fob", Action: "list", Description: "List Protect key fobs",
			Method: "GET", Path: "/v1/fobs", Example: "uictl fob list",
		},
		"fob.get": {
			Resource: "fob", Action: "get", Description: "Get Protect key fob details",
			Method: "GET", Path: "/v1/fobs/{id}", Example: "uictl fob get <id>",
		},
		"fob.update": {
			Resource: "fob", Action: "update", Description: "Update Protect key fob settings",
			Method: "PATCH", Path: "/v1/fobs/{id}", Mutating: true, DryRun: true,
			Example: `uictl fob update <id> --json-input '{"name":"Front Door"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"link-station.list": {
			Resource: "link-station", Action: "list", Description: "List Protect link stations",
			Method: "GET", Path: "/v1/link-stations", Example: "uictl link-station list",
		},
		"link-station.get": {
			Resource: "link-station", Action: "get", Description: "Get Protect link station details",
			Method: "GET", Path: "/v1/link-stations/{id}", Example: "uictl link-station get <id>",
		},
		"link-station.update": {
			Resource: "link-station", Action: "update", Description: "Update Protect link station settings",
			Method: "PATCH", Path: "/v1/link-stations/{id}", Mutating: true, DryRun: true,
			Example: `uictl link-station update <id> --json-input '{"name":"Gate"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"pos.ingest": {
			Resource: "pos", Action: "ingest", Description: "Ingest a Protect point-of-sale transaction",
			Method: "POST", Path: "/v1/pos/cameras/{id}/transactions", Mutating: true, DryRun: true,
			Example: `uictl pos ingest <camera-id> --json-input '{"type":"sale","externalId":"txn-1","amount":12.5}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON body (type sale|refund, externalId, amount; optional currency, lineItems, location)"}},
		},
		"relay.list": {
			Resource: "relay", Action: "list", Description: "List Protect relays",
			Method: "GET", Path: "/v1/relays", Example: "uictl relay list",
		},
		"relay.get": {
			Resource: "relay", Action: "get", Description: "Get Protect relay details",
			Method: "GET", Path: "/v1/relays/{id}", Example: "uictl relay get <id>",
		},
		"relay.update": {
			Resource: "relay", Action: "update", Description: "Update Protect relay settings",
			Method: "PATCH", Path: "/v1/relays/{id}", Mutating: true, DryRun: true,
			Example: `uictl relay update <id> --json-input '{"name":"Gate"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"relay.activate": {
			Resource: "relay", Action: "activate", Description: "Activate a Protect relay output",
			Method: "POST", Path: "/v1/relays/{id}/outputs/{outputId}/activate", Mutating: true, DryRun: true,
			Example: "uictl relay activate <id> <output-id> --state on --pulse-duration 1000",
			Flags: []SchemaFlag{
				{Name: "state", Type: "string", Desc: "Output state: on or off; omit to toggle"},
				{Name: "pulse-duration", Type: "integer", Desc: "Auto-off duration in milliseconds"},
			},
		},
		"siren.list": {
			Resource: "siren", Action: "list", Description: "List Protect sirens",
			Method: "GET", Path: "/v1/sirens", Example: "uictl siren list",
		},
		"siren.get": {
			Resource: "siren", Action: "get", Description: "Get Protect siren details",
			Method: "GET", Path: "/v1/sirens/{id}", Example: "uictl siren get <id>",
		},
		"siren.update": {
			Resource: "siren", Action: "update", Description: "Update Protect siren settings",
			Method: "PATCH", Path: "/v1/sirens/{id}", Mutating: true, DryRun: true,
			Example: `uictl siren update <id> --json-input '{"name":"Hall"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"siren.play": {
			Resource: "siren", Action: "play", Description: "Play a Protect siren",
			Method: "POST", Path: "/v1/sirens/{id}/play", Mutating: true, DryRun: true,
			Example: "uictl siren play <id> --duration 10",
			Flags:   []SchemaFlag{{Name: "duration", Type: "integer", Desc: "Duration in seconds (5, 10, 20, or 30)"}},
		},
		"siren.stop": {
			Resource: "siren", Action: "stop", Description: "Stop a Protect siren",
			Method: "POST", Path: "/v1/sirens/{id}/stop", Mutating: true, DryRun: true,
			Example: "uictl siren stop <id>",
		},
		"siren.test-sound": {
			Resource: "siren", Action: "test-sound", Description: "Play a Protect siren test sound",
			Method: "POST", Path: "/v1/sirens/{id}/test-sound", Mutating: true, DryRun: true,
			Example: "uictl siren test-sound <id> --volume 50",
			Flags:   []SchemaFlag{{Name: "volume", Type: "integer", Desc: "Volume from 1 to 100"}},
		},
		"speaker.list": {
			Resource: "speaker", Action: "list", Description: "List Protect speakers",
			Method: "GET", Path: "/v1/speakers", Example: "uictl speaker list",
		},
		"speaker.get": {
			Resource: "speaker", Action: "get", Description: "Get Protect speaker details",
			Method: "GET", Path: "/v1/speakers/{id}", Example: "uictl speaker get <id>",
		},
		"speaker.update": {
			Resource: "speaker", Action: "update", Description: "Update Protect speaker settings",
			Method: "PATCH", Path: "/v1/speakers/{id}", Mutating: true, DryRun: true,
			Example: `uictl speaker update <id> --json-input '{"name":"Porch"}'`,
			Flags:   []SchemaFlag{{Name: "json-input", Type: "string", Required: true, Desc: "JSON patch body"}},
		},
		"speaker.test-sound": {
			Resource: "speaker", Action: "test-sound", Description: "Play a Protect speaker test sound",
			Method: "POST", Path: "/v1/speakers/{id}/test-sound", Mutating: true, DryRun: true,
			Example: "uictl speaker test-sound <id> --volume 40",
			Flags:   []SchemaFlag{{Name: "volume", Type: "integer", Desc: "Volume from 0 to 100"}},
		},
		"subscribe.devices": {
			Resource: "subscribe", Action: "devices", Description: "Subscribe to Protect device websocket updates",
			Method: "GET", Path: "/v1/subscribe/devices", Example: "uictl subscribe devices",
		},
		"subscribe.events": {
			Resource: "subscribe", Action: "events", Description: "Subscribe to Protect event websocket messages",
			Method: "GET", Path: "/v1/subscribe/events", Example: "uictl subscribe events",
		},
		"ulp-user.list": {
			Resource: "ulp-user", Action: "list", Description: "List UniFi Identity users",
			Method: "GET", Path: "/v1/ulp-users", Example: "uictl ulp-user list",
		},
		"ulp-user.get": {
			Resource: "ulp-user", Action: "get", Description: "Get a UniFi Identity user",
			Method: "GET", Path: "/v1/ulp-users/{id}", Example: "uictl ulp-user get <id>",
		},
		"protect-user.list": {
			Resource: "protect-user", Action: "list", Description: "List Protect users",
			Method: "GET", Path: "/v1/users", Example: "uictl protect-user list",
		},
		"protect-user.get": {
			Resource: "protect-user", Action: "get", Description: "Get a Protect user",
			Method: "GET", Path: "/v1/users/{id}", Example: "uictl protect-user get <id>",
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
