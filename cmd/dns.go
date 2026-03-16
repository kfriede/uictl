package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dnsCmd)
}

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Manage DNS policies",
	Long: `Manage DNS policies within a site.

Examples:
  uictl dns list
  uictl dns get <policy-id>
  uictl dns create --json-input '{...}'
  uictl dns delete <policy-id> --yes`,
}

func init() {
	dnsCmd.AddCommand(dnsListCmd)
	dnsCmd.AddCommand(dnsGetCmd)
	dnsCmd.AddCommand(dnsCreateCmd)
	dnsCmd.AddCommand(dnsUpdateCmd)
	dnsCmd.AddCommand(dnsDeleteCmd)
}

var dnsListCmd = &cobra.Command{
	Use: "list", Short: "List DNS policies", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/dns/policies", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var dnsGetCmd = &cobra.Command{
	Use: "get <policy-id>", Short: "Get DNS policy", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/dns/policies/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var dnsCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a DNS policy", Args: cobra.NoArgs,
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
			printer.Status("[dry-run] Would create DNS policy:")
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/dns/policies", siteId), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success("Created DNS policy")
		return printAPIResult(result)
	},
}

func init() {
	dnsCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var dnsUpdateCmd = &cobra.Command{
	Use: "update <policy-id>", Short: "Update a DNS policy", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update DNS policy %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Put(fmt.Sprintf("/v1/sites/%s/dns/policies/%s", siteId, args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated DNS policy %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	dnsUpdateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var dnsDeleteCmd = &cobra.Command{
	Use: "delete <policy-id>", Short: "Delete a DNS policy", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete DNS policy %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete DNS policy %s", args[0])) {
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
		if err := client.Delete(fmt.Sprintf("/v1/sites/%s/dns/policies/%s", siteId, args[0])); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted DNS policy %s", args[0]))
		return nil
	},
}
