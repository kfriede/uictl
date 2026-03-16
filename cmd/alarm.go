package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(alarmCmd)
	rootCmd.AddCommand(protectInfoCmd)
}

var alarmCmd = &cobra.Command{
	Use:   "alarm",
	Short: "Manage Protect alarm webhooks",
}

func init() {
	alarmCmd.AddCommand(alarmWebhookCmd)
}

var alarmWebhookCmd = &cobra.Command{
	Use:   "webhook <trigger-id>",
	Short: "Send webhook to alarm manager",
	Long: `Send a webhook to trigger configured alarms.

The trigger-id is a user-defined string matching the alarm's configured trigger ID.

Examples:
  uictl alarm webhook my-alarm-trigger`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would send alarm webhook with trigger ID %s", args[0]))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var body any
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput != "" {
			b, err := parseJSONInput(jsonInput)
			if err != nil {
				return err
			}
			body = b
		}

		_, err = client.Post(fmt.Sprintf("/v1/alarm-manager/webhook/%s", args[0]), body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Sent alarm webhook %s", args[0]))
		return nil
	},
}

func init() {
	alarmWebhookCmd.Flags().String("json-input", "", "Optional JSON webhook payload")
}

var protectInfoCmd = &cobra.Command{
	Use:   "protect-info",
	Short: "Get Protect application info",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		data, err := client.Get("/v1/meta/info")
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}
