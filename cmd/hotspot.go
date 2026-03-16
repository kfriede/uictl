package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hotspotCmd)
}

var hotspotCmd = &cobra.Command{
	Use:   "hotspot",
	Short: "Manage hotspot vouchers",
	Long: `Generate, list, and delete hotspot vouchers for guest access.

Examples:
  uictl hotspot list
  uictl hotspot get <voucher-id>
  uictl hotspot create --name "Day Pass" --duration 1440 --count 10
  uictl hotspot delete <voucher-id> --yes`,
}

func init() {
	hotspotCmd.AddCommand(hotspotListCmd)
	hotspotCmd.AddCommand(hotspotGetCmd)
	hotspotCmd.AddCommand(hotspotCreateCmd)
	hotspotCmd.AddCommand(hotspotDeleteCmd)
}

var hotspotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List vouchers",
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
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/hotspot/vouchers", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var hotspotGetCmd = &cobra.Command{
	Use:   "get <voucher-id>",
	Short: "Get voucher details",
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/hotspot/vouchers/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var hotspotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate vouchers",
	Long: `Generate hotspot vouchers.

Examples:
  uictl hotspot create --name "Day Pass" --duration 1440
  uictl hotspot create --name "Event" --duration 480 --count 50 --limit 2048 --download 10000
  uictl hotspot create --json-input '{"name":"VIP","timeLimitMinutes":1440,"count":5}'`,
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
			duration, _ := cmd.Flags().GetInt("duration")
			if name == "" || duration == 0 {
				return fmt.Errorf("--name and --duration are required (or use --json-input)")
			}
			body = map[string]any{
				"name":             name,
				"timeLimitMinutes": duration,
			}
			if count, _ := cmd.Flags().GetInt("count"); count > 0 {
				body["count"] = count
			}
			if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
				body["dataUsageLimitMBytes"] = limit
			}
			if dl, _ := cmd.Flags().GetInt("download"); dl > 0 {
				body["rxRateLimitKbps"] = dl
			}
			if ul, _ := cmd.Flags().GetInt("upload"); ul > 0 {
				body["txRateLimitKbps"] = ul
			}
			if guests, _ := cmd.Flags().GetInt("guests"); guests > 0 {
				body["authorizedGuestLimit"] = guests
			}
		}

		if flagDryRun {
			printer.Status("[dry-run] Would create vouchers:")
			return printAPIResult(body)
		}

		resp, err := client.Post(fmt.Sprintf("/v1/sites/%s/hotspot/vouchers", siteId), body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success("Created vouchers")
		return printAPIResult(result)
	},
}

func init() {
	hotspotCreateCmd.Flags().String("name", "", "Voucher name")
	hotspotCreateCmd.Flags().Int("duration", 0, "Duration in minutes")
	hotspotCreateCmd.Flags().Int("count", 1, "Number of vouchers to generate")
	hotspotCreateCmd.Flags().Int("limit", 0, "Data cap in MB")
	hotspotCreateCmd.Flags().Int("download", 0, "Download rate limit in Kbps")
	hotspotCreateCmd.Flags().Int("upload", 0, "Upload rate limit in Kbps")
	hotspotCreateCmd.Flags().Int("guests", 0, "Max concurrent guests per voucher")
	hotspotCreateCmd.Flags().String("json-input", "", "Full JSON request body")
}

var hotspotDeleteCmd = &cobra.Command{
	Use:   "delete <voucher-id>",
	Short: "Delete a voucher",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete voucher %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete voucher %s", args[0])) {
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
		if err := client.Delete(fmt.Sprintf("/v1/sites/%s/hotspot/vouchers/%s", siteId, args[0])); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted voucher %s", args[0]))
		return nil
	},
}
