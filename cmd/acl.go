package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aclCmd)
}

var aclCmd = &cobra.Command{
	Use:   "acl",
	Short: "Manage ACL rules",
	Long: `Create, list, and manage Access Control List rules.

Examples:
  uictl acl list
  uictl acl get <rule-id>
  uictl acl create --json-input '{...}'
  uictl acl delete <rule-id> --yes
  uictl acl order`,
}

func init() {
	aclCmd.AddCommand(aclListCmd)
	aclCmd.AddCommand(aclGetCmd)
	aclCmd.AddCommand(aclCreateCmd)
	aclCmd.AddCommand(aclUpdateCmd)
	aclCmd.AddCommand(aclDeleteCmd)
	aclCmd.AddCommand(aclOrderCmd)
}

var aclListCmd = &cobra.Command{
	Use: "list", Short: "List ACL rules", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/acl-rules", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var aclGetCmd = &cobra.Command{
	Use: "get <rule-id>", Short: "Get ACL rule", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var aclCreateCmd = &cobra.Command{
	Use: "create", Short: "Create an ACL rule", Args: cobra.NoArgs,
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
			printer.Status("[dry-run] Would create ACL rule:")
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/acl-rules", siteId), body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success("Created ACL rule")
		return printAPIResult(result)
	},
}

func init() {
	aclCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var aclUpdateCmd = &cobra.Command{
	Use: "update <rule-id>", Short: "Update an ACL rule", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update ACL rule %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Put(fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteId, args[0]), body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Updated ACL rule %s", args[0]))
		return printAPIResult(result)
	},
}

func init() {
	aclUpdateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var aclDeleteCmd = &cobra.Command{
	Use: "delete <rule-id>", Short: "Delete an ACL rule", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete ACL rule %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete ACL rule %s", args[0])) {
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
		if err := client.Delete(fmt.Sprintf("/v1/sites/%s/acl-rules/%s", siteId, args[0])); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted ACL rule %s", args[0]))
		return nil
	},
}

var aclOrderCmd = &cobra.Command{
	Use:   "order",
	Short: "Get or set ACL rule ordering",
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
		path := fmt.Sprintf("/v1/sites/%s/acl-rules/ordering", siteId)

		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput != "" {
			body, err := parseJSONInput(jsonInput)
			if err != nil {
				return err
			}
			if flagDryRun {
				printer.Status("[dry-run] Would update ACL rule ordering:")
				return printAPIResult(body)
			}
			resp, err := client.Put(path, body)
			if err != nil {
				return err
			}
			var result map[string]any
			if err := json.Unmarshal(resp, &result); err != nil {
				return err
			}
			printer.Success("Updated ACL rule ordering")
			return printAPIResult(result)
		}

		var result map[string]any
		if err := client.GetJSON(path, &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

func init() {
	aclOrderCmd.Flags().String("json-input", "", "JSON ordering payload")
}
