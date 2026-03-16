package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(firewallCmd)
}

var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Manage firewall zones and policies",
	Long: `Manage custom firewall zones and policies within a site.

Examples:
  uictl firewall zone list
  uictl firewall policy list
  uictl firewall policy create --json-input '{...}'
  uictl firewall policy delete <id> --yes`,
}

// --- Zones ---

func init() {
	firewallCmd.AddCommand(firewallZoneCmd)
	firewallCmd.AddCommand(firewallPolicyCmd)
}

var firewallZoneCmd = &cobra.Command{
	Use:   "zone",
	Short: "Manage firewall zones",
}

func init() {
	firewallZoneCmd.AddCommand(fwZoneListCmd)
	firewallZoneCmd.AddCommand(fwZoneGetCmd)
	firewallZoneCmd.AddCommand(fwZoneCreateCmd)
	firewallZoneCmd.AddCommand(fwZoneUpdateCmd)
	firewallZoneCmd.AddCommand(fwZoneDeleteCmd)
}

var fwZoneListCmd = &cobra.Command{
	Use: "list", Short: "List firewall zones", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/firewall/zones", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var fwZoneGetCmd = &cobra.Command{
	Use: "get <zone-id>", Short: "Get firewall zone", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var fwZoneCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a firewall zone", Args: cobra.NoArgs,
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
			printer.Status("[dry-run] Would create firewall zone:")
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/firewall/zones", siteId), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success("Created firewall zone")
		return printAPIResult(result)
	},
}

func init() {
	fwZoneCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var fwZoneUpdateCmd = &cobra.Command{
	Use: "update <zone-id>", Short: "Update a firewall zone", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update firewall zone %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Put(fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteId, args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated firewall zone %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	fwZoneUpdateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var fwZoneDeleteCmd = &cobra.Command{
	Use: "delete <zone-id>", Short: "Delete a firewall zone", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete firewall zone %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete firewall zone %s", args[0])) {
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
		if err := client.Delete(fmt.Sprintf("/v1/sites/%s/firewall/zones/%s", siteId, args[0])); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted firewall zone %s", args[0]))
		return nil
	},
}

// --- Policies ---

var firewallPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage firewall policies",
}

func init() {
	firewallPolicyCmd.AddCommand(fwPolicyListCmd)
	firewallPolicyCmd.AddCommand(fwPolicyGetCmd)
	firewallPolicyCmd.AddCommand(fwPolicyCreateCmd)
	firewallPolicyCmd.AddCommand(fwPolicyUpdateCmd)
	firewallPolicyCmd.AddCommand(fwPolicyDeleteCmd)
	firewallPolicyCmd.AddCommand(fwPolicyOrderCmd)
}

var fwPolicyListCmd = &cobra.Command{
	Use: "list", Short: "List firewall policies", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/firewall/policies", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var fwPolicyGetCmd = &cobra.Command{
	Use: "get <policy-id>", Short: "Get firewall policy", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var fwPolicyCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a firewall policy", Args: cobra.NoArgs,
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
			printer.Status("[dry-run] Would create firewall policy:")
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/firewall/policies", siteId), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success("Created firewall policy")
		return printAPIResult(result)
	},
}

func init() {
	fwPolicyCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var fwPolicyUpdateCmd = &cobra.Command{
	Use: "update <policy-id>", Short: "Update a firewall policy", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update firewall policy %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Put(fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteId, args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		json.Unmarshal(resp, &result)
		printer.Success(fmt.Sprintf("Updated firewall policy %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	fwPolicyUpdateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var fwPolicyDeleteCmd = &cobra.Command{
	Use: "delete <policy-id>", Short: "Delete a firewall policy", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete firewall policy %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete firewall policy %s", args[0])) {
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
		if err := client.Delete(fmt.Sprintf("/v1/sites/%s/firewall/policies/%s", siteId, args[0])); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted firewall policy %s", args[0]))
		return nil
	},
}

var fwPolicyOrderCmd = &cobra.Command{
	Use:   "order",
	Short: "Get or set firewall policy ordering",
	Long: `Get or set the ordering of user-defined firewall policies.

Examples:
  uictl firewall policy order --source <zone-id> --dest <zone-id>
  uictl firewall policy order --source <zone-id> --dest <zone-id> --json-input '{"orderedPolicyIds":["id1","id2"]}'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		srcZone, _ := cmd.Flags().GetString("source")
		dstZone, _ := cmd.Flags().GetString("dest")
		if srcZone == "" || dstZone == "" {
			return fmt.Errorf("--source and --dest zone IDs are required")
		}
		path := fmt.Sprintf("/v1/sites/%s/firewall/policies/ordering?sourceFirewallZoneId=%s&destinationFirewallZoneId=%s", siteId, srcZone, dstZone)

		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput != "" {
			// Set ordering
			body, err := parseJSONInput(jsonInput)
			if err != nil {
				return err
			}
			if flagDryRun {
				printer.Status("[dry-run] Would update firewall policy ordering:")
				return printAPIResult(body)
			}
			resp, err := client.Put(path, body)
			if err != nil {
				return err
			}
			var result map[string]any
			json.Unmarshal(resp, &result)
			printer.Success("Updated firewall policy ordering")
			return printAPIResult(result)
		}

		// Get ordering
		var result map[string]any
		if err := client.GetJSON(path, &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

func init() {
	fwPolicyOrderCmd.Flags().String("source", "", "Source firewall zone ID")
	fwPolicyOrderCmd.Flags().String("dest", "", "Destination firewall zone ID")
	fwPolicyOrderCmd.Flags().String("json-input", "", "JSON ordering payload")
}
