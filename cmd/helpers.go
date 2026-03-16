package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kfriede/uictl/internal/api"
	"github.com/kfriede/uictl/internal/config"
	"github.com/kfriede/uictl/internal/output"
)

// newAPIClient creates an API client from the current configuration.
func newAPIClient() (*api.Client, error) {
	if cfg.Host == "" {
		printer.PrintError(output.NewError(
			output.ErrCodeConfig,
			"No controller host configured",
			"Run `uictl login` to configure your controller, or set UICTL_HOST.",
		))
		return nil, fmt.Errorf("no host configured")
	}

	apiKey := cfg.APIKey
	if apiKey == "" {
		// Try keyring
		secret, err := config.GetSecret(cfg.Profile)
		if err != nil {
			printer.PrintError(output.NewError(
				output.ErrCodeAuth,
				"Could not retrieve API key from keyring: "+err.Error(),
				"Run `uictl login` to re-authenticate.",
			))
			return nil, err
		}
		apiKey = secret
	}

	if apiKey == "" {
		printer.PrintError(output.NewAuthError("No API key configured"))
		return nil, fmt.Errorf("no API key configured")
	}

	return api.NewClient(api.ClientConfig{
		Host:      cfg.Host,
		APIKey:    apiKey,
		Insecure:  cfg.Insecure,
		Verbose:   cfg.Verbose,
		Debug:     cfg.Debug,
		ErrWriter: os.Stderr,
	}), nil
}

// requireSite returns the site ID, erroring if not set.
func requireSite() (string, error) {
	if cfg.Site == "" {
		printer.PrintError(output.NewError(
			output.ErrCodeConfig,
			"No site specified",
			"Use --site flag or set UICTL_SITE, or run `uictl config set site <name>`.",
		))
		return "", fmt.Errorf("no site specified")
	}
	return cfg.Site, nil
}

// confirmAction asks the user to confirm a destructive action.
// Returns true if confirmed or --yes was passed.
func confirmAction(action string) bool {
	if flagYes {
		return true
	}

	fmt.Fprintf(os.Stderr, "Are you sure you want to %s? (y/N): ", action)

	var response string
	_, _ = fmt.Scanln(&response)
	return response == "y" || response == "yes" || response == "Y"
}

// parseJSONInput parses the --json-input flag value into a map.
func parseJSONInput(jsonStr string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return result, nil
}

// printAPIResult handles the common pattern of printing API results
// with proper error handling and exit codes.
func printAPIResult(data any) error {
	return printer.PrintResult(data)
}
