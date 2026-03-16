package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cameraCmd)
}

var cameraCmd = &cobra.Command{
	Use:   "camera",
	Short: "Manage Protect cameras",
	Long: `View, configure, and control UniFi Protect cameras.

Examples:
  uictl camera list
  uictl camera get <camera-id>
  uictl camera update <camera-id> --json-input '{"name":"Front Door"}'
  uictl camera snapshot <camera-id> > snapshot.jpg
  uictl camera stream <camera-id>
  uictl camera ptz goto <camera-id> <slot>`,
}

func init() {
	cameraCmd.AddCommand(cameraListCmd)
	cameraCmd.AddCommand(cameraGetCmd)
	cameraCmd.AddCommand(cameraUpdateCmd)
	cameraCmd.AddCommand(cameraSnapshotCmd)
	cameraCmd.AddCommand(cameraDisableMicCmd)
	cameraCmd.AddCommand(cameraStreamCmd)
	cameraCmd.AddCommand(cameraStreamCreateCmd)
	cameraCmd.AddCommand(cameraStreamDeleteCmd)
	cameraCmd.AddCommand(cameraTalkbackCmd)
	cameraCmd.AddCommand(cameraPtzCmd)
}

var cameraListCmd = &cobra.Command{
	Use: "list", Short: "List all cameras", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/cameras", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var cameraGetCmd = &cobra.Command{
	Use: "get <camera-id>", Short: "Get camera details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/cameras/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var cameraUpdateCmd = &cobra.Command{
	Use: "update <camera-id>", Short: "Update camera settings", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput == "" {
			return fmt.Errorf("--json-input is required")
		}
		body, err := parseJSONInput(jsonInput)
		if err != nil {
			return err
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would update camera %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/cameras/%s", args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Updated camera %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	cameraUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

var cameraSnapshotCmd = &cobra.Command{
	Use:   "snapshot <camera-id>",
	Short: "Get a camera snapshot (outputs JPEG to stdout)",
	Long: `Get a snapshot image from a camera. Output is binary JPEG to stdout.

Examples:
  uictl camera snapshot <camera-id> > snapshot.jpg
  uictl camera snapshot <camera-id> --high-quality > snapshot_hq.jpg`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/v1/cameras/%s/snapshot", args[0])
		hq, _ := cmd.Flags().GetBool("high-quality")
		if hq {
			path += "?highQuality=true"
		}
		data, err := client.Get(path)
		if err != nil {
			return err
		}
		_, err = cmd.OutOrStdout().Write(data)
		return err
	},
}

func init() {
	cameraSnapshotCmd.Flags().Bool("high-quality", false, "Force 1080P or higher resolution")
}

var cameraDisableMicCmd = &cobra.Command{
	Use:   "disable-mic <camera-id>",
	Short: "Permanently disable camera microphone (IRREVERSIBLE)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would permanently disable mic on camera %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("PERMANENTLY disable microphone on camera %s (cannot be undone without factory reset)", args[0])) {
			printer.Status("Cancelled.")
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/cameras/%s/disable-mic-permanently", args[0]), nil)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Permanently disabled microphone on camera %s", args[0]))
		return printAPIResult(result)
	},
}

var cameraStreamCmd = &cobra.Command{
	Use: "stream <camera-id>", Short: "Get existing RTSPS stream URLs", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/cameras/%s/rtsps-stream", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var cameraStreamCreateCmd = &cobra.Command{
	Use:   "stream-create <camera-id>",
	Short: "Create RTSPS streams for quality levels",
	Long: `Create RTSPS streams for specified quality levels.

Examples:
  uictl camera stream-create <camera-id> --json-input '{"qualities":["high","medium","low"]}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput == "" {
			return fmt.Errorf("--json-input is required (e.g. '{\"qualities\":[\"high\",\"medium\",\"low\"]}')")
		}
		body, err := parseJSONInput(jsonInput)
		if err != nil {
			return err
		}
		if flagDryRun {
			printer.Status("[dry-run] Would create RTSPS streams:")
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/cameras/%s/rtsps-stream", args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success("Created RTSPS streams")
		return printAPIResult(result)
	},
}

func init() {
	cameraStreamCreateCmd.Flags().String("json-input", "", "JSON request body")
}

var cameraStreamDeleteCmd = &cobra.Command{
	Use:   "stream-delete <camera-id>",
	Short: "Delete camera RTSPS stream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete RTSPS stream for camera %s", args[0]))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		qualities, _ := cmd.Flags().GetString("qualities")
		if qualities == "" {
			return fmt.Errorf("--qualities is required (e.g. 'high,medium,low')")
		}
		path := fmt.Sprintf("/v1/cameras/%s/rtsps-stream?qualities=%s", args[0], qualities)
		if err := client.Delete(path); err != nil {
			return err
		}
		printer.Success("Deleted RTSPS stream")
		return nil
	},
}

func init() {
	cameraStreamDeleteCmd.Flags().String("qualities", "", "Comma-separated quality levels to delete")
}

var cameraTalkbackCmd = &cobra.Command{
	Use: "talkback <camera-id>", Short: "Create talkback session", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/cameras/%s/talkback-session", args[0]), nil)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

// --- PTZ ---

var cameraPtzCmd = &cobra.Command{
	Use:   "ptz",
	Short: "PTZ camera control",
}

func init() {
	cameraPtzCmd.AddCommand(ptzGotoCmd)
	cameraPtzCmd.AddCommand(ptzPatrolStartCmd)
	cameraPtzCmd.AddCommand(ptzPatrolStopCmd)
}

var ptzGotoCmd = &cobra.Command{
	Use: "goto <camera-id> <slot>", Short: "Move PTZ camera to preset (slot 0-4)", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would move camera %s to preset %s", args[0], args[1]))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		_, err = client.Post(fmt.Sprintf("/v1/cameras/%s/ptz/goto/%s", args[0], args[1]), nil)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Moved camera %s to preset %s", args[0], args[1]))
		return nil
	},
}

var ptzPatrolStartCmd = &cobra.Command{
	Use: "patrol-start <camera-id> <slot>", Short: "Start PTZ patrol (slot 0-4)", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would start patrol %s on camera %s", args[1], args[0]))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		_, err = client.Post(fmt.Sprintf("/v1/cameras/%s/ptz/patrol/start/%s", args[0], args[1]), nil)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Started patrol %s on camera %s", args[1], args[0]))
		return nil
	},
}

var ptzPatrolStopCmd = &cobra.Command{
	Use: "patrol-stop <camera-id>", Short: "Stop active PTZ patrol", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would stop patrol on camera %s", args[0]))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		_, err = client.Post(fmt.Sprintf("/v1/cameras/%s/ptz/patrol/stop", args[0]), nil)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Stopped patrol on camera %s", args[0]))
		return nil
	},
}
