package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ulpUserCmd)
}

var ulpUserCmd = &cobra.Command{
	Use:   "ulp-user",
	Short: "View UniFi Identity users",
	Long: `Read UniFi Identity (ULP) users known to Protect.

Examples:
  uictl ulp-user list
  uictl ulp-user get <id>`,
}

func init() {
	ulpUserCmd.AddCommand(ulpUserListCmd)
	ulpUserCmd.AddCommand(ulpUserGetCmd)
}

var ulpUserListCmd = &cobra.Command{
	Use: "list", Short: "List UniFi Identity users", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/ulp-users", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var ulpUserGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get a UniFi Identity user", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/ulp-users/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}
