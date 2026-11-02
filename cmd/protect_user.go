package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(protectUserCmd)
}

var protectUserCmd = &cobra.Command{
	Use:   "protect-user",
	Short: "View Protect users",
	Long: `Read UniFi Protect local users.

Examples:
  uictl protect-user list
  uictl protect-user get <id>`,
}

func init() {
	protectUserCmd.AddCommand(protectUserListCmd)
	protectUserCmd.AddCommand(protectUserGetCmd)
}

var protectUserListCmd = &cobra.Command{
	Use: "list", Short: "List Protect users", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/users", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var protectUserGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get a Protect user", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/users/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}
