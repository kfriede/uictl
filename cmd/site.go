package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(siteCmd)
}

var siteCmd = &cobra.Command{
	Use:   "site",
	Short: "Manage UniFi sites",
	Long: `List and manage UniFi sites.

Examples:
  uictl site list
  uictl site list --json --fields id,name`,
}

func init() {
	siteCmd.AddCommand(siteListCmd)
}

var siteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List local sites",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		data, err := client.GetAllPages("/v1/sites")
		if err != nil {
			return err
		}

		return printAPIResult(toAnySlice(data))
	},
}

// toAnySlice converts []map[string]any to []any for the printer.
func toAnySlice(data []map[string]any) []any {
	result := make([]any, len(data))
	for i, item := range data {
		result[i] = item
	}
	return result
}
