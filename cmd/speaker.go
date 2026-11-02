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
	Short: "Manage Protect speakers",
	Long: `View and control UniFi Protect speakers.

Examples:
  uictl speaker list
  uictl speaker get <id>
  uictl speaker update <id> --json-input '{"name":"Porch"}'
  uictl speaker test-sound <id> --volume 40`,
}

func init() {
	speakerCmd.AddCommand(speakerListCmd)
	speakerCmd.AddCommand(speakerGetCmd)
	speakerCmd.AddCommand(speakerUpdateCmd)
	speakerCmd.AddCommand(speakerTestSoundCmd)
}

var speakerListCmd = &cobra.Command{
	Use: "list", Short: "List all speakers", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/speakers", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var speakerGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get speaker details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/speakers/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var speakerUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update speaker settings", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update speaker %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/speakers/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated speaker %s", args[0]))
	},
}

func init() {
	speakerUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

var speakerTestSoundCmd = &cobra.Command{
	Use:   "test-sound <id>",
	Short: "Play a speaker test sound",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if cmd.Flags().Changed("volume") {
			volume, _ := cmd.Flags().GetInt("volume")
			if volume < 0 || volume > 100 {
				return fmt.Errorf("--volume must be between 0 and 100")
			}
			body["volume"] = volume
		}
		var payload any
		if len(body) > 0 {
			payload = body
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would test speaker %s sound:", args[0]))
			if payload != nil {
				return printAPIResult(payload)
			}
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/speakers/%s/test-sound", args[0]), payload)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Tested speaker %s sound", args[0]))
	},
}

func init() {
	speakerTestSoundCmd.Flags().Int("volume", 0, "Test volume from 0 to 100")
}
