package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(relayCmd)
}

var relayCmd = &cobra.Command{
	Use:   "relay",
	Short: "Manage Protect relays",
	Long: `View and control UniFi Protect relays.

Examples:
  uictl relay list
  uictl relay get <id>
  uictl relay update <id> --json-input '{"name":"Gate"}'
  uictl relay activate <id> <output-id> --state on --pulse-duration 1000`,
}

func init() {
	relayCmd.AddCommand(relayListCmd)
	relayCmd.AddCommand(relayGetCmd)
	relayCmd.AddCommand(relayUpdateCmd)
	relayCmd.AddCommand(relayActivateCmd)
}

var relayListCmd = &cobra.Command{
	Use: "list", Short: "List all relays", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/relays", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var relayGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get relay details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/relays/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var relayUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update relay settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update relay %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/relays/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated relay %s", args[0]))
	},
}

func init() {
	relayUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

var relayActivateCmd = &cobra.Command{
	Use:   "activate <id> <output-id>",
	Short: "Activate a relay output",
	Long: `Set, clear, or toggle a relay output.

Omit --state to toggle. --pulse-duration is milliseconds and applies when state is on.
An empty body toggles the output.

Examples:
  uictl relay activate <id> 0
  uictl relay activate <id> 0 --state on --pulse-duration 1000
  uictl relay activate <id> 0 --state off`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			if state != "on" && state != "off" {
				return fmt.Errorf("--state must be on or off")
			}
			body["state"] = state
		}
		if cmd.Flags().Changed("pulse-duration") {
			pulse, _ := cmd.Flags().GetInt("pulse-duration")
			if pulse < 0 {
				return fmt.Errorf("--pulse-duration must be >= 0")
			}
			body["pulseDuration"] = pulse
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would activate relay %s output %s:", args[0], args[1]))
			if payload != nil {
				return printAPIResult(payload)
			}
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/relays/%s/outputs/%s/activate", args[0], args[1]), payload)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Activated relay %s output %s", args[0], args[1]))
	},
}

func init() {
	relayActivateCmd.Flags().String("state", "", "Output state: on or off (omit to toggle)")
	relayActivateCmd.Flags().Int("pulse-duration", 0, "Auto-off duration in milliseconds when state is on")
}
