package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var externalArmProfileIDRe = regexp.MustCompile(`^[0-9A-Za-z-]{1,64}$`)

func validateExternalArmProfileID(id string) error {
	if !externalArmProfileIDRe.MatchString(id) {
		return fmt.Errorf("invalid external arm profile id %q (must match ^[0-9A-Za-z-]{1,64}$)", id)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(externalArmProfileCmd)
}

var externalArmProfileCmd = &cobra.Command{
	Use:   "external-arm-profile",
	Short: "Manage Protect external arm profiles",
	Long: `List, create, and arm UniFi Protect external alarm manager profiles.

External arm profiles are only available when the external alarm manager is in use.
They are distinct from local arm profiles.

Examples:
  uictl external-arm-profile list
  uictl external-arm-profile get <id>
  uictl external-arm-profile create --json-input '{"title":"Away"}'
  uictl external-arm-profile update <id> --json-input '{"title":"Night"}'
  uictl external-arm-profile delete <id> --yes
  uictl external-arm-profile arm <id>
  uictl external-arm-profile disarm <id>`,
}

func init() {
	externalArmProfileCmd.AddCommand(externalArmProfileListCmd)
	externalArmProfileCmd.AddCommand(externalArmProfileGetCmd)
	externalArmProfileCmd.AddCommand(externalArmProfileCreateCmd)
	externalArmProfileCmd.AddCommand(externalArmProfileUpdateCmd)
	externalArmProfileCmd.AddCommand(externalArmProfileDeleteCmd)
	externalArmProfileCmd.AddCommand(externalArmProfileArmCmd)
	externalArmProfileCmd.AddCommand(externalArmProfileDisarmCmd)
}

var externalArmProfileListCmd = &cobra.Command{
	Use: "list", Short: "List external arm profiles", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON("/v1/external/arm-profiles", &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var externalArmProfileGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get an external arm profile", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := validateExternalArmProfileID(id); err != nil {
			return err
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/external/arm-profiles/%s", id), &result); err != nil {
			return err
		}
		return printAPIResult(result)
	},
}

var externalArmProfileCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an external arm profile",
	Long: `Create an external arm profile. The JSON body requires title (non-empty string).

Optional fields: activation_delay, deactivation_delay, schedule, alarm_ids, pin_code.

Examples:
  uictl external-arm-profile create --json-input '{"title":"Away"}'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonInput, _ := cmd.Flags().GetString("json-input")
		if jsonInput == "" {
			return fmt.Errorf("--json-input is required")
		}
		body, err := parseJSONInput(jsonInput)
		if err != nil {
			return err
		}
		title, _ := body["title"].(string)
		if strings.TrimSpace(title) == "" {
			return fmt.Errorf("title is required and must be a non-empty string")
		}
		if flagDryRun {
			printer.Status("[dry-run] Would create external arm profile:")
			return printAPIResult(body)
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post("/v1/external/arm-profiles", body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success("Created external arm profile")
		return printAPIResult(result)
	},
}

func init() {
	externalArmProfileCreateCmd.Flags().String("json-input", "", "JSON request body (required; title is required)")
}

var externalArmProfileUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update an external arm profile", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := validateExternalArmProfileID(id); err != nil {
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
			printer.Status(fmt.Sprintf("[dry-run] Would update external arm profile %s:", id))
			return printAPIResult(body)
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Patch(fmt.Sprintf("/v1/external/arm-profiles/%s", id), body)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Updated external arm profile %s", id))
		return printAPIResult(result)
	},
}

func init() {
	externalArmProfileUpdateCmd.Flags().String("json-input", "", "JSON patch body")
}

var externalArmProfileDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete an external arm profile", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := validateExternalArmProfileID(id); err != nil {
			return err
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would delete external arm profile %s", id))
			return nil
		}
		if !confirmAction(fmt.Sprintf("delete external arm profile %s", id)) {
			printer.Status("Cancelled.")
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		if err := client.Delete(fmt.Sprintf("/v1/external/arm-profiles/%s", id)); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted external arm profile %s", id))
		return nil
	},
}

var externalArmProfileArmCmd = &cobra.Command{
	Use: "arm <id>", Short: "Arm an external arm profile", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := validateExternalArmProfileID(id); err != nil {
			return err
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would arm external arm profile %s", id))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/external/arm-profiles/%s/actions/arm", id), nil)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Armed external arm profile %s", id))
		return printAPIResult(result)
	},
}

var externalArmProfileDisarmCmd = &cobra.Command{
	Use: "disarm <id>", Short: "Disarm an external arm profile", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if err := validateExternalArmProfileID(id); err != nil {
			return err
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would disarm external arm profile %s", id))
			return nil
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.Post(fmt.Sprintf("/v1/external/arm-profiles/%s/actions/disarm", id), nil)
		if err != nil {
			return err
		}
		var result map[string]any
		if err := json.Unmarshal(resp, &result); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Disarmed external arm profile %s", id))
		return printAPIResult(result)
	},
}
