package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(speakerCmd)
}

var speakerCmd = &cobra.Command{
	Use:   "speaker",
	Short: "Control Protect speakers",
	Long: `Control UniFi Protect speakers.

Examples:
  uictl speaker stop-siren <id>`,
}

func init() {
	speakerCmd.AddCommand(speakerStopSirenCmd)
}

var speakerStopSirenCmd = &cobra.Command{
	Use:   "stop-siren <id>",
	Short: "Stop a speaker siren",
	Long: `Stop the siren on a Protect speaker.

Returns 400 if the speaker is offline or the firmware is too old.

Examples:
  uictl speaker stop-siren <id>
  uictl speaker stop-siren <id> --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if id == "" {
			return fmt.Errorf("speaker id is required")
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would stop siren on speaker %s", id))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		_, err = client.Post(fmt.Sprintf("/v1/speakers/%s/stop-siren", id), nil)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Stopped siren on speaker %s", id))
		return nil
	},
}
