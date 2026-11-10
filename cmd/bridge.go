package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(bridgeCmd)
}

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Manage Protect bridges",
	Long: `View and update UniFi Protect bridges.

Examples:
  uictl bridge list
  uictl bridge get <id>
  uictl bridge update <id> --json-input '{"name":"Garage"}'`,
}

func init() {
	bridgeCmd.AddCommand(bridgeListCmd)
	bridgeCmd.AddCommand(bridgeGetCmd)
	bridgeCmd.AddCommand(bridgeUpdateCmd)
}

var bridgeListCmd = &cobra.Command{
	Use: "list", Short: "List all bridges", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/bridges", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var bridgeGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get bridge details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/bridges/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var bridgeUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update bridge settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update bridge %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/bridges/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated bridge %s", args[0]))
	},
}

func init() {
	bridgeUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}
