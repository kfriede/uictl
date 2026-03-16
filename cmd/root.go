package cmd

import (
	"fmt"
	"os"

	"github.com/kfriede/uictl/internal/config"
	"github.com/kfriede/uictl/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	cfg     *config.Config
	printer *output.Printer

	// Global flags
	flagJSON     bool
	flagCSV      bool
	flagOutput   string
	flagFields   string
	flagQuiet    bool
	flagVerbose  bool
	flagDebug    bool
	flagNoColor  bool
	flagInsecure bool
	flagProfile  string
	flagSite     string
	flagYes      bool
	flagDryRun   bool
)

var rootCmd = &cobra.Command{
	Use:   "uictl",
	Short: "CLI for managing UniFi network devices and controllers",
	Long: `uictl is a command-line interface for managing Ubiquiti UniFi
network devices and controllers. Built for both humans and LLM agents.

Get started:
  uictl login                      Authenticate with your controller
  uictl site list                  List available sites
  uictl device list                List adopted devices
  uictl client list                List connected clients
  uictl network list               List networks

Use --json for machine-readable output, --fields to select specific fields.
Run 'uictl schema <resource>.<action>' for command introspection.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		// Print error to stderr (Cobra's SilenceErrors suppresses its own printing)
		if printer != nil {
			printer.PrintError(output.AppError{
				Code:    output.ErrCodeGeneral,
				Message: err.Error(),
			})
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	}
	return err
}

func init() {
	cobra.OnInitialize(initConfig, initPrinter)

	// Output flags
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagCSV, "csv", false, "Output as CSV")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table, json, csv, ndjson")
	rootCmd.PersistentFlags().StringVar(&flagFields, "fields", "", "Comma-separated list of fields to include")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")

	// Verbosity flags
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose logging to stderr")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug logging (full request/response bodies)")

	// Connection flags
	rootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "k", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().StringVarP(&flagProfile, "profile", "p", "", "Configuration profile to use")
	rootCmd.PersistentFlags().StringVarP(&flagSite, "site", "s", "", "Site name or ID")

	// Safety flags
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "Preview changes without executing")

	// Bind env vars
	viper.SetEnvPrefix("UICTL")
	_ = viper.BindEnv("host")
	_ = viper.BindEnv("api_key")
	_ = viper.BindEnv("site")
	_ = viper.BindEnv("profile")
	_ = viper.BindEnv("output_format")
	_ = viper.BindEnv("debug")
	_ = viper.BindEnv("no_color")

	// Register subcommands
	rootCmd.AddCommand(versionCmd)

	// Enable "did you mean?" suggestions
	EnableSuggestions()
}

func initConfig() {
	var err error
	cfg, err = config.Load(flagProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = config.Default()
	}

	// Apply flag overrides
	if flagSite != "" {
		cfg.Site = flagSite
	} else if s := viper.GetString("site"); s != "" {
		cfg.Site = s
	}

	if flagInsecure {
		cfg.Insecure = true
	}

	if flagDebug || viper.GetBool("debug") {
		cfg.Debug = true
		cfg.Verbose = true
	} else if flagVerbose {
		cfg.Verbose = true
	}
}

func initPrinter() {
	format := resolveOutputFormat()
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	noColor := flagNoColor || viper.GetBool("no_color") || os.Getenv("NO_COLOR") != ""

	printer = output.NewPrinter(output.PrinterConfig{
		Format:  format,
		IsTTY:   isTTY,
		NoColor: noColor,
		Quiet:   flagQuiet,
		Fields:  flagFields,
		Writer:  os.Stdout,
		ErrWriter: os.Stderr,
	})
}

// resolveOutputFormat determines the output format from flags, env, and TTY detection.
// Precedence: --json/--csv flags > --output flag > UICTL_OUTPUT_FORMAT env > TTY detection
func resolveOutputFormat() output.Format {
	if flagJSON {
		return output.FormatJSON
	}
	if flagCSV {
		return output.FormatCSV
	}
	if flagOutput != "" {
		switch flagOutput {
		case "json":
			return output.FormatJSON
		case "csv":
			return output.FormatCSV
		case "ndjson":
			return output.FormatNDJSON
		case "table":
			return output.FormatTable
		default:
			fmt.Fprintf(os.Stderr, "Warning: unknown output format %q, using default\n", flagOutput)
		}
	}
	if envFmt := viper.GetString("output_format"); envFmt != "" {
		switch envFmt {
		case "json":
			return output.FormatJSON
		case "csv":
			return output.FormatCSV
		case "ndjson":
			return output.FormatNDJSON
		case "table":
			return output.FormatTable
		}
	}

	// Auto-detect: non-TTY defaults to JSON (LLM-friendly)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return output.FormatJSON
	}
	return output.FormatTable
}
