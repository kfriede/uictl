package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(wifiCmd)
}

var wifiCmd = &cobra.Command{
	Use:   "wifi",
	Short: "Manage WiFi broadcasts (SSIDs)",
	Long: `Create, update, delete, and inspect WiFi broadcast configurations.

Examples:
  uictl wifi list
  uictl wifi get <broadcast-id>
  uictl wifi create --json-input '{"name":"Guest WiFi",...}'
  uictl wifi delete <broadcast-id> --yes`,
}

func init() {
	wifiCmd.AddCommand(wifiListCmd)
	wifiCmd.AddCommand(wifiGetCmd)
	wifiCmd.AddCommand(wifiCreateCmd)
	wifiCmd.AddCommand(wifiUpdateCmd)
	wifiCmd.AddCommand(wifiDeleteCmd)
}

var wifiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List WiFi broadcasts",
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
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/wifi/broadcasts", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var wifiGetCmd = &cobra.Command{
	Use:   "get <broadcast-id>",
	Short: "Get WiFi broadcast details",
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var wifiCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a WiFi broadcast",
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
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput == "" {
			return fmt.Errorf("--json-input is required")
		}
		body, err := parseJSONInput(jsonInput)
		if err != nil {
			return err
		}
		if flagDryRun {
			printer.Status("[dry-run] Would create WiFi broadcast:")
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/wifi/broadcasts", siteId), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success("Created WiFi broadcast")
		return printAPIResult(result)
	},
}

func init() {
	wifiCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var wifiUpdateCmd = &cobra.Command{
	Use:   "update <broadcast-id>",
	Short: "Update a WiFi broadcast",
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
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput == "" {
			return fmt.Errorf("--json-input is required")
		}
		body, err := parseJSONInput(jsonInput)
		if err != nil {
			return err
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would update WiFi broadcast %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Put(fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteId, args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated WiFi broadcast %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	wifiUpdateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var wifiDeleteCmd = &cobra.Command{
	Use:   "delete <broadcast-id>",
	Short: "Delete a WiFi broadcast",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete WiFi broadcast %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete WiFi broadcast %s", args[0])) {
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
		forceFlag, _ := cmd.Flags().GetBool("force")
		path := fmt.Sprintf("/v1/sites/%s/wifi/broadcasts/%s", siteId, args[0])
		if forceFlag {
			path += "?force=true"
		}
		if err := client.Delete(path); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted WiFi broadcast %s", args[0]))
		return nil
	},
}

func init() {
	wifiDeleteCmd.Flags().Bool("force", false, "Force delete even if in use")
}
