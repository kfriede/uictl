package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(posCmd)
}

var posCmd = &cobra.Command{
	Use:   "pos",
	Short: "Ingest Protect point-of-sale transactions",
	Long: `Send a point-of-sale transaction to a Protect camera.

The JSON body requires type (sale or refund), externalId, and amount.
Optional fields: currency, lineItems, location.

Examples:
  uictl pos ingest <camera-id> --json-input '{"type":"sale","externalId":"txn-1","amount":12.5,"currency":"USD"}'`,
}

func init() {
	posCmd.AddCommand(posIngestCmd)
}

var posIngestCmd = &cobra.Command{
	Use:   "ingest <camera-id>",
	Short: "Ingest a POS transaction for a camera",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
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
			printer.Status(fmt.Sprintf("[dry-run] Would ingest POS transaction for camera %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Post(fmt.Sprintf("/v1/pos/cameras/%s/transactions", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Ingested POS transaction for camera %s", args[0]))
	},
}

func init() {
	posIngestCmd.Flags().String("json-input", "", "JSON body (type sale|refund, externalId, amount; optional currency, lineItems, location)")
}
