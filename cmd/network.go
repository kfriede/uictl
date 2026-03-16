package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(networkCmd)
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage networks",
	Long: `Create, update, delete, and inspect network configurations.

Examples:
  uictl network list
  uictl network get <network-id>
  uictl network create --name "IoT" --vlan 30
  uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'
  uictl network delete <network-id> --yes
  uictl network references <network-id>`,
}

func init() {
	networkCmd.AddCommand(networkListCmd)
	networkCmd.AddCommand(networkGetCmd)
	networkCmd.AddCommand(networkCreateCmd)
	networkCmd.AddCommand(networkUpdateCmd)
	networkCmd.AddCommand(networkDeleteCmd)
	networkCmd.AddCommand(networkRefsCmd)
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List networks",
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

		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/networks", siteId))
		if err != nil {
			return err
		}

		return printAPIResult(toAnySlice(data))
	},
}

var networkGetCmd = &cobra.Command{
	Use:   "get <network-id>",
	Short: "Get network details",
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/networks/%s", siteId, args[0]), &result); err != nil {
			return err
		}

		return printAPIResult(result)
	},
}

var networkCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a network",
	Long: `Create a new network.

Examples:
  uictl network create --name "IoT" --vlan 30
  uictl network create --json-input '{"name":"IoT","enabled":true,"management":false,"vlanId":30}'`,
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

		var body map[string]any
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput != "" {
			body, err = parseJSONInput(jsonInput)
			if err != nil {
				return err
			}
		} else {
			name, _ := cmd.Flags().GetString("name")
			vlan, _ := cmd.Flags().GetInt("vlan")
			if name == "" {
				return fmt.Errorf("--name or --json-input is required")
			}
			body = map[string]any{
				"name":       name,
				"enabled":    true,
				"management": false,
				"vlanId":     vlan,
			}
		}

		if flagDryRun {
			printer.Status("[dry-run] Would create network:")
			return printAPIResult(body)
		}

		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/networks", siteId), body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Created network %q", body["name"]))
		return printAPIResult(result)
	},
}

func init() {
	networkCreateCmd.Flags().String("name", "", "Network name")
	networkCreateCmd.Flags().Int("vlan", 0, "VLAN ID")
	networkCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var networkUpdateCmd = &cobra.Command{
	Use:   "update <network-id>",
	Short: "Update a network",
	Long: `Update an existing network.

Examples:
  uictl network update <id> --json-input '{"name":"IoT v2","enabled":true,"management":false,"vlanId":30}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		networkId := args[0]

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
			return fmt.Errorf("--json-input is required for updates")
		}

		body, err := parseJSONInput(jsonInput)
		if err != nil {
			return err
		}

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would update network %s:", networkId))
			return printAPIResult(body)
		}

		resp, err := client.Put(fmt.Sprintf("/v1/sites/%s/networks/%s", siteId, networkId), body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Updated network %s", networkId))
		return printAPIResult(result)
	},
}

func init() {
	networkUpdateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var networkDeleteCmd = &cobra.Command{
	Use:   "delete <network-id>",
	Short: "Delete a network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		networkId := args[0]

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete network %s", networkId))
			return nil
		}

		if !confirmAction(fmt.Sprintf("delete network %s", networkId)) {
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
		path := fmt.Sprintf("/v1/sites/%s/networks/%s", siteId, networkId)
		if forceFlag {
			path += "?force=true"
		}

		if err := client.Delete(path); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Deleted network %s", networkId))
		return nil
	},
}

func init() {
	networkDeleteCmd.Flags().Bool("force", false, "Force delete even if in use")
}

var networkRefsCmd = &cobra.Command{
	Use:   "references <network-id>",
	Short: "Get network references",
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/networks/%s/references", siteId, args[0]), &result); err != nil {
			return err
		}

		return printAPIResult(result)
	},
}
