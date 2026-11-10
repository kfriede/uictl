package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(linkStationCmd)
}

var linkStationCmd = &cobra.Command{
	Use:   "link-station",
	Short: "Manage Protect link stations",
	Long: `View and update UniFi Protect link stations.

Examples:
  uictl link-station list
  uictl link-station get <id>
  uictl link-station update <id> --json-input '{"name":"Gate"}'`,
}

func init() {
	linkStationCmd.AddCommand(linkStationListCmd)
	linkStationCmd.AddCommand(linkStationGetCmd)
	linkStationCmd.AddCommand(linkStationUpdateCmd)
}

var linkStationListCmd = &cobra.Command{
	Use: "list", Short: "List all link stations", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/link-stations", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var linkStationGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get link station details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/link-stations/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var linkStationUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update link station settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update link station %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/link-stations/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated link station %s", args[0]))
	},
}

func init() {
	linkStationUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}
