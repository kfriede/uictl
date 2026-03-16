package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kfriede/uictl/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a UniFi controller",
	Long: `Authenticate with a UniFi controller using an API key.

Examples:
  uictl login                        Interactive login
  uictl login --host 192.168.1.1     Specify host
  uictl login --profile office       Save as named profile

The API key is stored in your OS keyring when available,
falling back to the config file with restrictive permissions.`,
	Args: cobra.NoArgs,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().String("host", "", "Controller hostname or IP address")
}

func runLogin(cmd *cobra.Command, args []string) error {
	host, _ := cmd.Flags().GetString("host")

	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	if !isTTY {
		return fmt.Errorf("login requires an interactive terminal; set UICTL_HOST and UICTL_API_KEY environment variables for non-interactive use")
	}

	reader := bufio.NewReader(os.Stdin)

	// Get host
	if host == "" && cfg.Host != "" {
		fmt.Fprintf(os.Stderr, "Controller host [%s]: ", cfg.Host)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			host = input
		} else {
			host = cfg.Host
		}
	} else if host == "" {
		fmt.Fprint(os.Stderr, "Controller host (e.g. 192.168.1.1 or unifi.local): ")
		input, _ := reader.ReadString('\n')
		host = strings.TrimSpace(input)
		if host == "" {
			return fmt.Errorf("host is required")
		}
	}

	// Get API key
	fmt.Fprint(os.Stderr, "API Key: ")
	apiKeyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after hidden input
	if err != nil {
		return fmt.Errorf("reading API key: %w", err)
	}
	apiKey := strings.TrimSpace(string(apiKeyBytes))
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Get site
	site := cfg.Site
	fmt.Fprintf(os.Stderr, "Default site [%s]: ", site)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		site = input
	}

	// Ask about TLS
	insecure := cfg.Insecure
	fmt.Fprint(os.Stderr, "Skip TLS verification? (y/N): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "y" || input == "yes" {
		insecure = true
	}

	// Store API key in keyring
	profile := flagProfile
	if config.KeyringAvailable() {
		if err := config.StoreSecret(profile, apiKey); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not store API key in keyring: %v\n", err)
			fmt.Fprintln(os.Stderr, "API key will be stored in the config file instead.")
			// Fall through to save in config
		} else {
			printer.Status("API key stored in OS keyring")
			apiKey = "" // Don't save in config file
		}
	} else {
		printer.Status("OS keyring not available, storing API key in config file")
	}

	// Save config
	newCfg := &config.Config{
		Host:     host,
		Site:     site,
		Insecure: insecure,
		Profile:  profile,
	}
	if apiKey != "" {
		newCfg.APIKey = apiKey
	}

	if err := config.Save(newCfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	printer.Success(fmt.Sprintf("Logged in to %s (site: %s)", host, site))
	printer.Status("Run `uictl site list` to verify your connection.")
	return nil
}
