package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(alarmHubCmd)
}

var alarmHubCmd = &cobra.Command{
	Use:   "alarm-hub",
	Short: "Manage Protect alarm hubs",
	Long: `View and control UniFi Protect alarm hubs.

Examples:
  uictl alarm-hub list
  uictl alarm-hub get <id>
  uictl alarm-hub update <id> --json-input '{"name":"Warehouse"}'
  uictl alarm-hub trigger <id> <output-id> --enable --duration 5000`,
}

func init() {
	alarmHubCmd.AddCommand(alarmHubListCmd)
	alarmHubCmd.AddCommand(alarmHubGetCmd)
	alarmHubCmd.AddCommand(alarmHubUpdateCmd)
	alarmHubCmd.AddCommand(alarmHubTriggerCmd)
}

var alarmHubListCmd = &cobra.Command{
	Use: "list", Short: "List all alarm hubs", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/alarm-hubs", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var alarmHubGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get alarm hub details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/alarm-hubs/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var alarmHubUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update alarm hub settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update alarm hub %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/alarm-hubs/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated alarm hub %s", args[0]))
	},
}

func init() {
	alarmHubUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

var alarmHubTriggerCmd = &cobra.Command{
	Use:   "trigger <id> <output-id>",
	Short: "Trigger an alarm hub output",
	Long: `Trigger an alarm hub output channel.

Omit --enable to toggle. --delay and --duration are milliseconds.
An empty body toggles the output.

Examples:
  uictl alarm-hub trigger <id> 0
  uictl alarm-hub trigger <id> 0 --enable=false
  uictl alarm-hub trigger <id> 0 --enable --delay 0 --duration 5000`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if cmd.Flags().Changed("enable") {
			enabled, _ := cmd.Flags().GetBool("enable")
			body["enable"] = enabled
		}
		if cmd.Flags().Changed("delay") {
			delay, _ := cmd.Flags().GetInt("delay")
			if delay < 0 {
				return fmt.Errorf("--delay must be >= 0")
			}
			body["delay"] = delay
		}
		if cmd.Flags().Changed("duration") {
			duration, _ := cmd.Flags().GetInt("duration")
			if duration < 0 {
				return fmt.Errorf("--duration must be >= 0")
			}
			body["duration"] = duration
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would trigger alarm hub %s output %s:", args[0], args[1]))
			if payload != nil {
				return printAPIResult(payload)
			}
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/alarm-hubs/%s/outputs/%s/trigger", args[0], args[1]), payload)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Triggered alarm hub %s output %s", args[0], args[1]))
	},
}

func init() {
	alarmHubTriggerCmd.Flags().Bool("enable", false, "Set output on (true) or off (false); omit to toggle")
	alarmHubTriggerCmd.Flags().Int("delay", 0, "Delay in milliseconds before the output activates")
	alarmHubTriggerCmd.Flags().Int("duration", 0, "Duration in milliseconds to keep the output active (0 = until turned off)")
}

// mutationResult prints a success line. Empty and 204 bodies are not unmarshaled.
func mutationResult(resp []byte, success string) error {
	if len(resp) == 0 {
		printer.Success(success)
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	printer.Success(success)
	return printAPIResult(result)
}
