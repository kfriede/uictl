package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kfriede/uictl/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage configuration",
	Long: `View and manage uictl configuration.

Examples:
  uictl config show                 Show current configuration
  uictl config path                 Show config file path
  uictl config set host 10.0.0.1   Set a config value
  uictl config profiles             List saved profiles`,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configSetCmd)
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Redact API key
		display := map[string]any{
			"host":     cfg.Host,
			"site":     cfg.Site,
			"profile":  cfg.Profile,
			"insecure": cfg.Insecure,
			"verbose":  cfg.Verbose,
			"debug":    cfg.Debug,
		}

		// Check if API key is set (don't reveal it)
		if cfg.APIKey != "" {
			display["apiKey"] = "****" + cfg.APIKey[max(0, len(cfg.APIKey)-4):]
		} else {
			secret, _ := config.GetSecret(cfg.Profile)
			if secret != "" {
				display["apiKey"] = "(stored in keyring)"
			} else {
				display["apiKey"] = "(not set)"
			}
		}

		return printer.PrintResult(display)
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config directory path",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stdout, config.Dir())
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Supported keys: host, site, insecure

Examples:
  uictl config set host 192.168.1.1
  uictl config set site default
  uictl config set insecure true`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(args[0])
		value := args[1]

		switch key {
		case "host":
			cfg.Host = value
		case "site":
			cfg.Site = value
		case "insecure":
			cfg.Insecure = value == "true" || value == "1" || value == "yes"
		default:
			return fmt.Errorf("unknown config key: %s (supported: host, site, insecure)", key)
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		printer.Success(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	},
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// marshalJSON is a helper to output config as JSON, unused but available for future.
func marshalJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
