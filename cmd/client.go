package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Manage connected clients",
	Long: `View and manage connected clients (wired, wireless, VPN, guest).

Examples:
  uictl client list
  uictl client list --fields id,name,ipAddress,type
  uictl client get <client-id>
  uictl client authorize <client-id>
  uictl client unauthorize <client-id>`,
}

func init() {
	clientCmd.AddCommand(clientListCmd)
	clientCmd.AddCommand(clientGetCmd)
	clientCmd.AddCommand(clientAuthorizeCmd)
	clientCmd.AddCommand(clientUnauthorizeCmd)
}

var clientListCmd = &cobra.Command{
	Use:   "list",
	Short: "List connected clients",
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

		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/clients", siteId))
		if err != nil {
			return err
		}

		return printAPIResult(toAnySlice(data))
	},
}

var clientGetCmd = &cobra.Command{
	Use:   "get <client-id>",
	Short: "Get connected client details",
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/clients/%s", siteId, args[0]), &result); err != nil {
			return err
		}

		return printAPIResult(result)
	},
}

var clientAuthorizeCmd = &cobra.Command{
	Use:   "authorize <client-id>",
	Short: "Authorize a guest client",
	Long: `Authorize network access for a guest client.

Examples:
  uictl client authorize <client-id>
  uictl client authorize <client-id> --json-input '{"action":"authorize","expiresAt":"2025-06-01T00:00:00Z"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientId := args[0]

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would authorize client %s", clientId))
			return nil
		}

		apiClient, err := newAPIClient()
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
			body = map[string]any{"action": "authorize"}
		}

		resp, err := apiClient.Post(fmt.Sprintf("/v1/sites/%s/clients/%s/actions", siteId, clientId), body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Authorized client %s", clientId))
		return printAPIResult(result)
	},
}

func init() {
	clientAuthorizeCmd.Flags().String("json-input", "", "Full JSON request body")
}

var clientUnauthorizeCmd = &cobra.Command{
	Use:   "unauthorize <client-id>",
	Short: "Unauthorize a guest client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientId := args[0]

		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would unauthorize client %s", clientId))
			return nil
		}

		apiClient, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}

		body := map[string]any{"action": "unauthorize"}
		resp, err := apiClient.Post(fmt.Sprintf("/v1/sites/%s/clients/%s/actions", siteId, clientId), body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Unauthorized client %s", clientId))
		return printAPIResult(result)
	},
}
