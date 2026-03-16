package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lightCmd)
	rootCmd.AddCommand(sensorCmd)
	rootCmd.AddCommand(chimeCmd)
	rootCmd.AddCommand(viewerCmd)
	rootCmd.AddCommand(liveviewCmd)
	rootCmd.AddCommand(nvrCmd)
}

// --- Lights ---

var lightCmd = &cobra.Command{
	Use:   "light",
	Short: "Manage Protect lights",
}

func init() {
	lightCmd.AddCommand(lightListCmd)
	lightCmd.AddCommand(lightGetCmd)
	lightCmd.AddCommand(lightUpdateCmd)
}

var lightListCmd = &cobra.Command{
	Use: "list", Short: "List all lights", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/lights", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var lightGetCmd = &cobra.Command{
	Use: "get <light-id>", Short: "Get light details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/lights/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var lightUpdateCmd = &cobra.Command{
	Use: "update <light-id>", Short: "Update light settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update light %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/lights/%s", args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated light %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	lightUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

// --- Sensors ---

var sensorCmd = &cobra.Command{
	Use:   "sensor",
	Short: "Manage Protect sensors",
}

func init() {
	sensorCmd.AddCommand(sensorListCmd)
	sensorCmd.AddCommand(sensorGetCmd)
	sensorCmd.AddCommand(sensorUpdateCmd)
}

var sensorListCmd = &cobra.Command{
	Use: "list", Short: "List all sensors", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/sensors", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var sensorGetCmd = &cobra.Command{
	Use: "get <sensor-id>", Short: "Get sensor details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/sensors/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var sensorUpdateCmd = &cobra.Command{
	Use: "update <sensor-id>", Short: "Update sensor settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update sensor %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/sensors/%s", args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated sensor %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	sensorUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

// --- Chimes ---

var chimeCmd = &cobra.Command{
	Use:   "chime",
	Short: "Manage Protect chimes",
}

func init() {
	chimeCmd.AddCommand(chimeListCmd)
	chimeCmd.AddCommand(chimeGetCmd)
	chimeCmd.AddCommand(chimeUpdateCmd)
}

var chimeListCmd = &cobra.Command{
	Use: "list", Short: "List all chimes", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/chimes", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var chimeGetCmd = &cobra.Command{
	Use: "get <chime-id>", Short: "Get chime details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/chimes/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var chimeUpdateCmd = &cobra.Command{
	Use: "update <chime-id>", Short: "Update chime settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update chime %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/chimes/%s", args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated chime %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	chimeUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

// --- Viewers ---

var viewerCmd = &cobra.Command{
	Use:   "viewer",
	Short: "Manage Protect viewers",
}

func init() {
	viewerCmd.AddCommand(viewerListCmd)
	viewerCmd.AddCommand(viewerGetCmd)
	viewerCmd.AddCommand(viewerUpdateCmd)
}

var viewerListCmd = &cobra.Command{
	Use: "list", Short: "List all viewers", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/viewers", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var viewerGetCmd = &cobra.Command{
	Use: "get <viewer-id>", Short: "Get viewer details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/viewers/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var viewerUpdateCmd = &cobra.Command{
	Use: "update <viewer-id>", Short: "Update viewer settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update viewer %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/viewers/%s", args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated viewer %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	viewerUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

// --- Live Views ---

var liveviewCmd = &cobra.Command{
	Use:   "liveview",
	Short: "Manage Protect live views",
}

func init() {
	liveviewCmd.AddCommand(liveviewListCmd)
	liveviewCmd.AddCommand(liveviewGetCmd)
	liveviewCmd.AddCommand(liveviewCreateCmd)
	liveviewCmd.AddCommand(liveviewUpdateCmd)
}

var liveviewListCmd = &cobra.Command{
	Use: "list", Short: "List all live views", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/liveviews", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var liveviewGetCmd = &cobra.Command{
	Use: "get <liveview-id>", Short: "Get live view details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/liveviews/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var liveviewCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a live view", Args: cobra.NoArgs,
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
			printer.Status("[dry-run] Would create live view:")
			return printAPIResult(body)
		}
		resp, err := client.Post("/v1/liveviews", body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success("Created live view")
		return printAPIResult(result)
	},
}

func init() {
	liveviewCreateCmd.Flags().String("json-input", "", "JSON request body")
}

var liveviewUpdateCmd = &cobra.Command{
	Use: "update <liveview-id>", Short: "Update live view configuration", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update live view %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/liveviews/%s", args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated live view %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	liveviewUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

// --- NVR ---

var nvrCmd = &cobra.Command{
	Use:   "nvr",
	Short: "View Protect NVR details",
}

func init() {
	nvrCmd.AddCommand(nvrGetCmd)
}

var nvrGetCmd = &cobra.Command{
	Use: "get", Short: "Get NVR details", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON("/v1/nvrs", &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}
