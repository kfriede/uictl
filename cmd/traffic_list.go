package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(trafficListCmd)
}

var trafficListCmd = &cobra.Command{
	Use:     "traffic-list",
	Aliases: []string{"tl"},
	Short:   "Manage traffic matching lists",
	Long: `Manage port and IP address lists used across firewall policy configurations.

Examples:
  uictl traffic-list list
  uictl traffic-list get <list-id>
  uictl traffic-list create --json-input '{...}'
  uictl traffic-list delete <list-id> --yes`,
}

func init() {
	trafficListCmd.AddCommand(tlListCmd)
	trafficListCmd.AddCommand(tlGetCmd)
	trafficListCmd.AddCommand(tlCreateCmd)
	trafficListCmd.AddCommand(tlUpdateCmd)
	trafficListCmd.AddCommand(tlDeleteCmd)
}

var tlListCmd = &cobra.Command{
	Use: "list", Short: "List traffic matching lists", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/traffic-matching-lists", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var tlGetCmd = &cobra.Command{
	Use: "get <list-id>", Short: "Get traffic matching list", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var tlCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a traffic matching list", Args: cobra.NoArgs,
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
			printer.Status("[dry-run] Would create traffic matching list:")
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/traffic-matching-lists", siteId), body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success("Created traffic matching list")
		return printAPIResult(result)
	},
}

func init() {
	tlCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var tlUpdateCmd = &cobra.Command{
	Use: "update <list-id>", Short: "Update a traffic matching list", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update traffic matching list %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Put(fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteId, args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Updated traffic matching list %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	tlUpdateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var tlDeleteCmd = &cobra.Command{
	Use: "delete <list-id>", Short: "Delete a traffic matching list", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete traffic matching list %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete traffic matching list %s", args[0])) {
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
		if err := client.Delete(fmt.Sprintf("/v1/sites/%s/traffic-matching-lists/%s", siteId, args[0])); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted traffic matching list %s", args[0]))
		return nil
	},
}
