package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(apiCmd)
}

var apiCmd = &cobra.Command{
	Use:   "api <method> <path> [--data <json>]",
	Short: "Make raw API requests",
	Long: `Make raw API requests to the UniFi controller.

This is a passthrough command for endpoints not yet wrapped by
dedicated subcommands. The path is relative to the integration API base.

Examples:
  uictl api get /v1/info
  uictl api get /v1/sites
  uictl api post /v1/sites/{siteId}/devices --data '{"macAddress":"aa:bb:cc:dd:ee:ff"}'
  uictl api delete /v1/sites/{siteId}/networks/{networkId}`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAPI,
}

func init() {
	apiCmd.Flags().StringP("data", "d", "", "JSON request body")
}

func runAPI(cmd *cobra.Command, args []string) error {
	method := strings.ToUpper(args[0])
	path := args[1]

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	switch method {
	case "GET":
		data, err := client.Get(path)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, string(data))

	case "POST", "PUT", "PATCH":
		dataFlag, _ := cmd.Flags().GetString("data")
		var body any
		if dataFlag != "" {
			body = rawJSON(dataFlag)
		} else if hasStdin() {
			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			body = rawJSON(string(stdinData))
		}

		var resp []byte
		switch method {
		case "POST":
			resp, err = client.Post(path, body)
		case "PUT":
			resp, err = client.Put(path, body)
		case "PATCH":
			resp, err = client.Patch(path, body)
		}
		if err != nil {
			return err
		}
		if resp != nil {
			_, _ = fmt.Fprintln(os.Stdout, string(resp))
		}

	case "DELETE":
		if err := client.Delete(path); err != nil {
			return err
		}
		printer.Success("Deleted " + path)

	default:
		return fmt.Errorf("unsupported HTTP method: %s (use GET, POST, PUT, PATCH, or DELETE)", method)
	}

	return nil
}

// rawJSON wraps a string so it marshals as raw JSON (not double-encoded).
type rawJSON string

func (r rawJSON) MarshalJSON() ([]byte, error) {
	return []byte(r), nil
}

func hasStdin() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
