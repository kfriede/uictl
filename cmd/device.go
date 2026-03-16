package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deviceCmd)
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Manage UniFi devices",
	Long: `List, inspect, and manage UniFi network devices.

Examples:
  uictl device list
  uictl device get <device-id>
  uictl device restart <device-id>
  uictl device adopt --mac aa:bb:cc:dd:ee:ff
  uictl device remove <device-id> --yes
  uictl device stats <device-id>`,
}

func init() {
	deviceCmd.AddCommand(deviceListCmd)
	deviceCmd.AddCommand(deviceGetCmd)
	deviceCmd.AddCommand(deviceAdoptCmd)
	deviceCmd.AddCommand(deviceRemoveCmd)
	deviceCmd.AddCommand(deviceActionCmd)
	deviceCmd.AddCommand(deviceStatsCmd)
	deviceCmd.AddCommand(devicePortActionCmd)
	deviceCmd.AddCommand(devicePendingCmd)
}

var deviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List adopted devices",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}

		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/devices", siteId))
		if err != nil {
			return err
		}

		return printAPIResult(toAnySlice(data))
	},
}

var deviceGetCmd = &cobra.Command{
	Use:   "get <device-id>",
	Short: "Get adopted device details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}

		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/devices/%s", siteId, args[0]), &result); err != nil {
			return err
		}

		return printAPIResult(result)
	},
}

var devicePendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List devices pending adoption",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		data, err := client.GetAllPages("/v1/pending-devices")
		if err != nil {
			return err
		}

		return printAPIResult(toAnySlice(data))
	},
}

var deviceAdoptCmd = &cobra.Command{
	Use:   "adopt",
	Short: "Adopt a device",
	Long: `Adopt a pending device by MAC address.

Examples:
  uictl device adopt --mac aa:bb:cc:dd:ee:ff
  uictl device adopt --json-input '{"macAddress":"aa:bb:cc:dd:ee:ff","ignoreDeviceLimit":false}'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status("[dry-run] Would adopt device")
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

		var body map[string]any
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput != "" {
			body, err = parseJSONInput(jsonInput)
			if err != nil {
				return err
			}
		} else {
			mac, _ := cmd.Flags().GetString("mac")
			if mac == "" {
				return fmt.Errorf("--mac or --json-input is required")
			}
			ignoreLimitFlag, _ := cmd.Flags().GetBool("ignore-limit")
			body = map[string]any{
				"macAddress":        mac,
				"ignoreDeviceLimit": ignoreLimitFlag,
			}
		}

		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/devices", siteId), body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Adopted device %s", body["macAddress"]))
		return printAPIResult(result)
	},
}

func init() {
	deviceAdoptCmd.Flags().String("mac", "", "MAC address of device to adopt")
	deviceAdoptCmd.Flags().Bool("ignore-limit", false, "Ignore device limit")
	deviceAdoptCmd.Flags().String("json-input", "", "Full JSON request body")
}

var deviceRemoveCmd = &cobra.Command{
	Use:   "remove <device-id>",
	Short: "Remove (unadopt) a device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceId := args[0]

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would remove device %s", deviceId))
			return nil
		}

		if !confirmAction(fmt.Sprintf("remove device %s", deviceId)) {
			printer.Status("Cancelled.")
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

		if err := client.Delete(fmt.Sprintf("/v1/sites/%s/devices/%s", siteId, deviceId)); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Removed device %s", deviceId))
		return nil
	},
}

var deviceActionCmd = &cobra.Command{
	Use:   "restart <device-id>",
	Short: "Restart a device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceId := args[0]

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would restart device %s", deviceId))
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

		body := map[string]any{"action": "restart"}
		_, err = client.Post(fmt.Sprintf("/v1/sites/%s/devices/%s/actions", siteId, deviceId), body)
		if err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Restarted device %s", deviceId))
		return nil
	},
}

var deviceStatsCmd = &cobra.Command{
	Use:   "stats <device-id>",
	Short: "Get latest device statistics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}

		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/devices/%s/statistics/latest", siteId, args[0]), &result); err != nil {
			return err
		}

		return printAPIResult(result)
	},
}

var devicePortActionCmd = &cobra.Command{
	Use:   "port-action <device-id> <port-idx> <action>",
	Short: "Execute a port action (e.g. cycle)",
	Args:  cobra.ExactArgs(3),
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
