package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	deviceCmd.AddCommand(devicePortCmd)
}

var devicePortCmd = &cobra.Command{
	Use:   "port",
	Short: "Manage switch ports",
	Long: `List and configure switch ports, including VLAN assignments.

These commands use the classic UniFi API under the hood, as per-port
VLAN configuration is not available through the integration API.

Examples:
  uictl device port list <device-id>
  uictl device port list <device-id> --fields idx,name,nativeNetwork,speed
  uictl device port set <device-id> <port-idx> --network VLAN80_Home`,
}

func init() {
	devicePortCmd.AddCommand(devicePortListCmd)
	devicePortCmd.AddCommand(devicePortSetCmd)
	devicePortCmd.AddCommand(devicePortCycleCmd)
}

// classicDeviceResponse wraps the classic API stat/device response.
type classicDeviceResponse struct {
	Data []classicDevice `json:"data"`
}

type classicDevice struct {
	ID            string         `json:"_id"`
	MAC           string         `json:"mac"`
	Name          string         `json:"name"`
	Model         string         `json:"model"`
	PortTable     []portEntry    `json:"port_table"`
	PortOverrides []portOverride `json:"port_overrides"`
}

type portEntry struct {
	PortIdx           int    `json:"port_idx"`
	Name              string `json:"name"`
	Media             string `json:"media"`
	Speed             int    `json:"speed"`
	FullDuplex        bool   `json:"full_duplex"`
	IsUplink          bool   `json:"is_uplink"`
	Up                bool   `json:"up"`
	Enable            bool   `json:"enable"`
	POEEnable         bool   `json:"poe_enable"`
	POEMode           string `json:"poe_mode"`
	NativeNetworkconf string `json:"native_networkconf_id"`
	OpMode            string `json:"op_mode"`
	PortconfID        string `json:"portconf_id"`
}

type portOverride struct {
	PortIdx             int    `json:"port_idx"`
	Name                string `json:"name,omitempty"`
	PortconfID          string `json:"portconf_id,omitempty"`
	NativeNetworkconfID string `json:"native_networkconf_id,omitempty"`
	SettingPreference   string `json:"setting_preference,omitempty"`
	POEMode             string `json:"poe_mode,omitempty"`
	OpMode              string `json:"op_mode,omitempty"`
	AggregateNumPorts   int    `json:"aggregate_num_ports,omitempty"`
}

// classicNetworkResponse wraps the classic API network list response.
type classicNetworkResponse struct {
	Data []classicNetwork `json:"data"`
}

type classicNetwork struct {
	ID      string `json:"_id"`
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
	VLAN    any    `json:"vlan"`    // can be string or number
	VLANId  any    `json:"vlan_id"` // alternate field
}

var devicePortListCmd = &cobra.Command{
	Use:   "list <device-id>",
	Short: "List switch ports with VLAN assignments",
	Long: `List all ports on a switch with their current VLAN/network assignments.

Accepts a device UUID (from the integration API), MAC address, or device name.

Examples:
  uictl device port list <device-id>
  uictl device port list <device-id> --fields idx,name,nativeNetwork,speed,poe,up`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceIdentifier := args[0]

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		siteId, err := requireSite()
		if err != nil {
			return err
		}
		siteRef, err := requireSiteRef()
		if err != nil {
			return err
		}

		// Get device MAC from integration API
		mac, err := resolveDeviceMAC(client, siteId, deviceIdentifier)
		if err != nil {
			return err
		}

		// Fetch device from classic API with port_table
		classicPath := fmt.Sprintf("/proxy/network/api/s/%s/stat/device/%s", siteRef, mac)
		var devResp classicDeviceResponse
		if err := client.GetRawJSON(classicPath, &devResp); err != nil {
			return fmt.Errorf("fetching port data from classic API: %w", err)
		}
		if len(devResp.Data) == 0 {
			return fmt.Errorf("device %s not found in classic API", deviceIdentifier)
		}

		dev := devResp.Data[0]

		// Fetch networks to resolve network IDs to names
		networkNames, err := getClassicNetworkNames(client, siteRef)
		if err != nil {
			// Non-fatal: show IDs instead of names
			networkNames = map[string]string{}
		}

		// Build port summary
		ports := make([]any, 0, len(dev.PortTable))
		for _, p := range dev.PortTable {
			nativeNetwork := resolveNetworkName(p.NativeNetworkconf, networkNames)

			// Find matching override for config vs runtime info
			portconfName := ""
			for _, o := range dev.PortOverrides {
				if o.PortIdx == p.PortIdx {
					if o.NativeNetworkconfID != "" {
						nativeNetwork = resolveNetworkName(o.NativeNetworkconfID, networkNames)
					}
					portconfName = o.PortconfID
					break
				}
			}

			port := map[string]any{
				"idx":           p.PortIdx,
				"name":          p.Name,
				"up":            p.Up,
				"speed":         p.Speed,
				"fullDuplex":    p.FullDuplex,
				"media":         p.Media,
				"isUplink":      p.IsUplink,
				"nativeNetwork": nativeNetwork,
				"portProfile":   portconfName,
				"poe":           p.POEEnable,
				"poeMode":       p.POEMode,
				"opMode":        p.OpMode,
			}
			ports = append(ports, port)
		}

		return printAPIResult(ports)
	},
}

var devicePortSetCmd = &cobra.Command{
	Use:   "set <device-id> <port-idx>",
	Short: "Configure a switch port",
	Long: `Set the native network (VLAN) on a switch port.

The --network flag accepts a network name, classic network ID, or VLAN ID.
This command performs a read-modify-write on the device's port_overrides.

Examples:
  uictl device port set <device-id> 4 --network VLAN80_Home
  uictl device port set <device-id> 4 --network VLAN80_Home --dry-run
  uictl device port set <device-id> 4 --json-input '{"native_networkconf_id":"<classic-id>","portconf_id":"All"}'`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceIdentifier := args[0]
		portIdx, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("port index must be a number: %w", err)
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		siteId, err := requireSite()
		if err != nil {
			return err
		}
		siteRef, err := requireSiteRef()
		if err != nil {
			return err
		}

		// Get device MAC from integration API
		mac, err := resolveDeviceMAC(client, siteId, deviceIdentifier)
		if err != nil {
			return err
		}

		// Fetch current device state from classic API
		classicPath := fmt.Sprintf("/proxy/network/api/s/%s/stat/device/%s", siteRef, mac)
		var devResp classicDeviceResponse
		if err := client.GetRawJSON(classicPath, &devResp); err != nil {
			return fmt.Errorf("fetching device from classic API: %w", err)
		}
		if len(devResp.Data) == 0 {
			return fmt.Errorf("device %s not found in classic API", deviceIdentifier)
		}

		dev := devResp.Data[0]

		// Determine the new override for this port
		var newOverride portOverride
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput != "" {
			if err := json.Unmarshal([]byte(jsonInput), &newOverride); err != nil {
				return fmt.Errorf("invalid JSON input: %w", err)
			}
			newOverride.PortIdx = portIdx
		} else {
			networkFlag, _ := cmd.Flags().GetString("network")
			if networkFlag == "" {
				return fmt.Errorf("--network or --json-input is required")
			}

			// Resolve network flag to classic network ID
			networkID, err := resolveClassicNetworkID(client, siteRef, networkFlag)
			if err != nil {
				return err
			}

			// Build override: preserve existing fields, update network
			newOverride = portOverride{PortIdx: portIdx}
			for _, o := range dev.PortOverrides {
				if o.PortIdx == portIdx {
					newOverride = o
					break
				}
			}
			newOverride.NativeNetworkconfID = networkID
			// "manual" tells the controller we're overriding the port profile
			newOverride.SettingPreference = "manual"
		}

		// Build the new port_overrides array (replace or append)
		overrides := make([]portOverride, 0, len(dev.PortOverrides)+1)
		found := false
		for _, o := range dev.PortOverrides {
			if o.PortIdx == portIdx {
				overrides = append(overrides, newOverride)
				found = true
			} else {
				overrides = append(overrides, o)
			}
		}
		if !found {
			overrides = append(overrides, newOverride)
		}

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would update port %d on device %s:", portIdx, deviceIdentifier))
			return printAPIResult(map[string]any{
				"deviceId":       dev.ID,
				"deviceName":     dev.Name,
				"portIdx":        portIdx,
				"newOverride":    newOverride,
				"totalOverrides": len(overrides),
			})
		}

		if !confirmAction(fmt.Sprintf("change port %d on device %s (%s)", portIdx, dev.Name, mac)) {
			printer.Status("Cancelled.")
			return nil
		}

		// PUT the updated port_overrides
		putPath := fmt.Sprintf("/proxy/network/api/s/%s/rest/device/%s", siteRef, dev.ID)
		body := map[string]any{"port_overrides": overrides}
		resp, err := client.PutRaw(putPath, body)
		if err != nil {
			return fmt.Errorf("updating port config: %w", err)
		}

		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			// PUT succeeded but response isn't JSON — still ok
			printer.Success(fmt.Sprintf("Updated port %d on device %s", portIdx, dev.Name))
			return nil
		}

		printer.Success(fmt.Sprintf("Updated port %d on device %s", portIdx, dev.Name))
		return nil
	},
}

func init() {
	devicePortSetCmd.Flags().String("network", "", "Network name, VLAN ID, or classic network ID")
	devicePortSetCmd.Flags().String("json-input", "", "Full JSON port override object")
}

var devicePortCycleCmd = &cobra.Command{
	Use:   "action <device-id> <port-idx> <action>",
	Short: "Execute a port action (e.g. cycle)",
	Long: `Execute an action on a specific port (e.g. power cycle PoE).

This is equivalent to 'uictl device port-action' but grouped under the port subcommand.

Examples:
  uictl device port action <device-id> 4 cycle`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceId := args[0]
		portIdx := args[1]
		action := args[2]

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would %s port %s on device %s", action, portIdx, deviceId))
			return nil
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}

		body := map[string]any{"action": action}
		_, err = client.Post(fmt.Sprintf("/v1/sites/%s/devices/%s/interfaces/ports/%s/actions", siteId, deviceId, portIdx), body)
		if err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Executed %s on port %s of device %s", action, portIdx, deviceId))
		return nil
	},
}

// resolveDeviceMAC fetches a device from the integration API and returns its MAC.
func resolveDeviceMAC(client interface {
	GetJSON(string, any) error
	GetAllPages(string) ([]map[string]any, error)
}, siteId, identifier string) (string, error) {
	// If it already looks like a MAC address, return it directly
	if isMAC(identifier) {
		return strings.ToLower(identifier), nil
	}

	// Try fetching by UUID first
	var device map[string]any
	err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/devices/%s", siteId, identifier), &device)
	if err == nil {
		if mac, ok := device["macAddress"].(string); ok {
			return strings.ToLower(mac), nil
		}
		if mac, ok := device["mac"].(string); ok {
			return strings.ToLower(mac), nil
		}
	}

	// Fall back to searching by name
	devices, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/devices", siteId))
	if err != nil {
		return "", fmt.Errorf("listing devices: %w", err)
	}

	for _, d := range devices {
		name, _ := d["name"].(string)
		if strings.EqualFold(name, identifier) {
			if mac, ok := d["macAddress"].(string); ok {
				return strings.ToLower(mac), nil
			}
			if mac, ok := d["mac"].(string); ok {
				return strings.ToLower(mac), nil
			}
		}
	}

	return "", fmt.Errorf("device %q not found; use a device UUID, MAC address, or name", identifier)
}

// isMAC checks if a string looks like a MAC address.
func isMAC(s string) bool {
	s = strings.ToLower(s)
	// Accept aa:bb:cc:dd:ee:ff or aabbccddeeff
	if len(s) == 17 && strings.Count(s, ":") == 5 {
		return true
	}
	if len(s) == 12 {
		for _, c := range s {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
		return true
	}
	return false
}

// getClassicNetworkNames returns a map of classic network ID → name.
func getClassicNetworkNames(client interface {
	GetRawJSON(string, any) error
}, siteRef string) (map[string]string, error) {
	var resp classicNetworkResponse
	path := fmt.Sprintf("/proxy/network/api/s/%s/rest/networkconf", siteRef)
	if err := client.GetRawJSON(path, &resp); err != nil {
		return nil, err
	}

	names := make(map[string]string, len(resp.Data))
	for _, n := range resp.Data {
		names[n.ID] = n.Name
	}
	return names, nil
}

// resolveClassicNetworkID resolves a network name, VLAN ID, or classic network ID
// to the classic API's network _id.
func resolveClassicNetworkID(client interface {
	GetRawJSON(string, any) error
}, siteRef, identifier string) (string, error) {
	var resp classicNetworkResponse
	path := fmt.Sprintf("/proxy/network/api/s/%s/rest/networkconf", siteRef)
	if err := client.GetRawJSON(path, &resp); err != nil {
		return "", fmt.Errorf("fetching networks: %w", err)
	}

	// Try exact ID match first
	for _, n := range resp.Data {
		if n.ID == identifier {
			return n.ID, nil
		}
	}

	// Try name match (case-insensitive)
	for _, n := range resp.Data {
		if strings.EqualFold(n.Name, identifier) {
			return n.ID, nil
		}
	}

	// Try VLAN ID match
	for _, n := range resp.Data {
		vlanStr := fmt.Sprintf("%v", n.VLAN)
		vlanIdStr := fmt.Sprintf("%v", n.VLANId)
		if vlanStr == identifier || vlanIdStr == identifier {
			return n.ID, nil
		}
	}

	// Build helpful error with available networks
	var available []string
	for _, n := range resp.Data {
		available = append(available, fmt.Sprintf("  %s (%s)", n.Name, n.ID))
	}
	return "", fmt.Errorf("network %q not found; available networks:\n%s", identifier, strings.Join(available, "\n"))
}

// resolveNetworkName resolves a classic network ID to its display name.
func resolveNetworkName(id string, names map[string]string) string {
	if id == "" {
		return ""
	}
	if name, ok := names[id]; ok {
		return name
	}
	return id
}
