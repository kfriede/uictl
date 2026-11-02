package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sirenCmd)
}

var sirenCmd = &cobra.Command{
	Use:   "siren",
	Short: "Manage Protect sirens",
	Long: `View and control UniFi Protect sirens.

Examples:
  uictl siren list
  uictl siren get <id>
  uictl siren update <id> --json-input '{"name":"Hall"}'
  uictl siren play <id> --duration 10
  uictl siren stop <id>
  uictl siren test-sound <id> --volume 50`,
}

func init() {
	sirenCmd.AddCommand(sirenListCmd)
	sirenCmd.AddCommand(sirenGetCmd)
	sirenCmd.AddCommand(sirenUpdateCmd)
	sirenCmd.AddCommand(sirenPlayCmd)
	sirenCmd.AddCommand(sirenStopCmd)
	sirenCmd.AddCommand(sirenTestSoundCmd)
}

var sirenListCmd = &cobra.Command{
	Use: "list", Short: "List all sirens", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/sirens", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var sirenGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get siren details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/sirens/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var sirenUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update siren settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update siren %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/sirens/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated siren %s", args[0]))
	},
}

func init() {
	sirenUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

var sirenPlayCmd = &cobra.Command{
	Use:   "play <id>",
	Short: "Play a siren",
	Long: `Play a siren. --duration is seconds. Allowed values: 5, 10, 20, 30. Omit to use the device default.

Examples:
  uictl siren play <id>
  uictl siren play <id> --duration 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if cmd.Flags().Changed("duration") {
			duration, _ := cmd.Flags().GetInt("duration")
			switch duration {
			case 5, 10, 20, 30:
				body["duration"] = duration
			default:
				return fmt.Errorf("--duration must be 5, 10, 20, or 30 seconds")
			}
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would play siren %s:", args[0]))
			if payload != nil {
				return printAPIResult(payload)
			}
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sirens/%s/play", args[0]), payload)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Played siren %s", args[0]))
	},
}

func init() {
	sirenPlayCmd.Flags().Int("duration", 0, "Play duration in seconds (5, 10, 20, or 30)")
}

var sirenStopCmd = &cobra.Command{
	Use: "stop <id>", Short: "Stop a siren", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would stop siren %s", args[0]))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sirens/%s/stop", args[0]), nil)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Stopped siren %s", args[0]))
	},
}

var sirenTestSoundCmd = &cobra.Command{
	Use:   "test-sound <id>",
	Short: "Play a siren test sound",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if cmd.Flags().Changed("volume") {
			volume, _ := cmd.Flags().GetInt("volume")
			if volume < 1 || volume > 100 {
				return fmt.Errorf("--volume must be between 1 and 100")
			}
			body["volume"] = volume
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would test siren %s sound:", args[0]))
			if payload != nil {
				return printAPIResult(payload)
			}
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sirens/%s/test-sound", args[0]), payload)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Tested siren %s sound", args[0]))
	},
}

func init() {
	sirenTestSoundCmd.Flags().Int("volume", 0, "Test volume from 1 to 100")
}
