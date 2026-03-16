package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(switchingCmd)
}

var switchingCmd = &cobra.Command{
	Use:   "switching",
	Short: "View switching features (read-only)",
	Long: `View switch stacking, LAG, and MC-LAG configurations.

Examples:
  uictl switching lag list
  uictl switching lag get <lag-id>
  uictl switching stack list
  uictl switching mc-lag list`,
}

func init() {
	switchingCmd.AddCommand(lagCmd)
	switchingCmd.AddCommand(stackCmd)
	switchingCmd.AddCommand(mcLagCmd)
}

// --- LAGs ---

var lagCmd = &cobra.Command{
	Use:   "lag",
	Short: "View LAG configurations",
}

func init() {
	lagCmd.AddCommand(lagListCmd)
	lagCmd.AddCommand(lagGetCmd)
}

var lagListCmd = &cobra.Command{
	Use: "list", Short: "List LAGs", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/switching/lags", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var lagGetCmd = &cobra.Command{
	Use: "get <lag-id>", Short: "Get LAG details", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/switching/lags/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

// --- Switch Stacks ---

var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "View switch stacks",
}

func init() {
	stackCmd.AddCommand(stackListCmd)
	stackCmd.AddCommand(stackGetCmd)
}

var stackListCmd = &cobra.Command{
	Use: "list", Short: "List switch stacks", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/switching/switch-stacks", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var stackGetCmd = &cobra.Command{
	Use: "get <stack-id>", Short: "Get switch stack", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/switching/switch-stacks/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

// --- MC-LAG Domains ---

var mcLagCmd = &cobra.Command{
	Use:   "mc-lag",
	Short: "View MC-LAG domains",
}

func init() {
	mcLagCmd.AddCommand(mcLagListCmd)
	mcLagCmd.AddCommand(mcLagGetCmd)
}

var mcLagListCmd = &cobra.Command{
	Use: "list", Short: "List MC-LAG domains", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/switching/mc-lag-domains", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var mcLagGetCmd = &cobra.Command{
	Use: "get <mc-lag-id>", Short: "Get MC-LAG domain", Args: cobra.ExactArgs(1),
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
		if err := client.GetJSON(fmt.Sprintf("/v1/sites/%s/switching/mc-lag-domains/%s", siteId, args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}
