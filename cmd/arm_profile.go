package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(armProfileCmd)
}

var armProfileCmd = &cobra.Command{
	Use:   "arm-profile",
	Short: "Manage Protect arm profiles",
	Long: `List, create, and switch UniFi Protect arm profiles.

Examples:
  uictl arm-profile list
  uictl arm-profile create --json-input '{"name":"Away","automations":[],"schedules":[],"recordEverything":false,"activationDelay":0}'
  uictl arm-profile update <id> --json-input '{"name":"Away"}'
  uictl arm-profile delete <id> --yes
  uictl arm-profile enable
  uictl arm-profile disable
  uictl arm-profile settings <arm-profile-id>`,
}

func init() {
	armProfileCmd.AddCommand(armProfileListCmd)
	armProfileCmd.AddCommand(armProfileCreateCmd)
	armProfileCmd.AddCommand(armProfileUpdateCmd)
	armProfileCmd.AddCommand(armProfileDeleteCmd)
	armProfileCmd.AddCommand(armProfileEnableCmd)
	armProfileCmd.AddCommand(armProfileDisableCmd)
	armProfileCmd.AddCommand(armProfileSettingsCmd)
}

var armProfileListCmd = &cobra.Command{
	Use: "list", Short: "List all arm profiles", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/arm-profiles", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var armProfileCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an arm profile",
	Long: `Create an arm profile. The JSON body requires name, automations, schedules, recordEverything, and activationDelay.

Examples:
  uictl arm-profile create --json-input '{"name":"Away","automations":[],"schedules":[],"recordEverything":false,"activationDelay":0}'`,
	Args: cobra.NoArgs,
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
			printer.Status("[dry-run] Would create arm profile:")
			return printAPIResult(body)
		}
		resp, err := client.Post("/v1/arm-profiles", body)
		if err != nil {
			return err
		}
		return mutationResult(resp, "Created arm profile")
	},
}

func init() {
	armProfileCreateCmd.Flags().String("json-input", "", "JSON body (name, automations, schedules, recordEverything, activationDelay)")
}

var armProfileUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update an arm profile", Args: cobra.ExactArgs(1),
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
			printer.Status(fmt.Sprintf("[dry-run] Would update arm profile %s:", args[0]))
			return printAPIResult(body)
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/arm-profiles/%s", args[0]), body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Updated arm profile %s", args[0]))
	},
}

func init() {
	armProfileUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

var armProfileDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete an arm profile", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete arm profile %s", args[0]))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete arm profile %s", args[0])) {
			printer.Status("Cancelled.")
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		if err := client.Delete(fmt.Sprintf("/v1/arm-profiles/%s", args[0])); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted arm profile %s", args[0]))
		return nil
	},
}

var armProfileEnableCmd = &cobra.Command{
	Use: "enable", Short: "Enable the arm alarm", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status("[dry-run] Would enable arm alarm")
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post("/v1/arm-profiles/enable", nil)
		if err != nil {
			return err
		}
		return mutationResult(resp, "Enabled arm alarm")
	},
}

var armProfileDisableCmd = &cobra.Command{
	Use: "disable", Short: "Disable the arm alarm", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDryRun {
			printer.Status("[dry-run] Would disable arm alarm")
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post("/v1/arm-profiles/disable", nil)
		if err != nil {
			return err
		}
		return mutationResult(resp, "Disabled arm alarm")
	},
}

var armProfileSettingsCmd = &cobra.Command{
	Use:   "settings <arm-profile-id>",
	Short: "Set the current arm profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{"armProfileId": args[0]}
		if flagDryRun {
			printer.Status("[dry-run] Would set current arm profile:")
			return printAPIResult(body)
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Patch("/v1/arm-profiles/settings", body)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Set current arm profile %s", args[0]))
	},
}
