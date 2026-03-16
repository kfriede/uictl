package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get controller application info",
	Long: `Get general details about the UniFi Network application.

Examples:
  uictl info
  uictl info --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var result map[string]any
		if err := client.GetJSON("/v1/info", &result); err != nil {
			return err
		}

		return printAPIResult(result)
	},
}
