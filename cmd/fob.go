package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fobCmd)
}

var fobCmd = &cobra.Command{
	Use:   "fob",
	Short: "Manage Protect key fobs",
	Long: `View and update UniFi Protect key fobs.

Examples:
  uictl fob list
  uictl fob get <id>
  uictl fob update <id> --json-input '{"name":"Front Door"}'`,
}

func init() {
	fobCmd.AddCommand(fobListCmd)
	fobCmd.AddCommand(fobGetCmd)
	fobCmd.AddCommand(fobUpdateCmd)
}

var fobListCmd = &cobra.Command{
	Use: "list", Short: "List all key fobs", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/fobs", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var fobGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get key fob details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/fobs/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var fobUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update key fob settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update fob %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/fobs/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated fob %s", args[0]))
	},
}

func init() {
	fobUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}
