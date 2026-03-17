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

Use --raw to bypass the /integration/ prefix and send the path directly
to the controller. This is required for classic API endpoints.

Examples:
  uictl api get /v1/info
  uictl api get /v1/sites
  uictl api post /v1/sites/{siteId}/devices --data '{"macAddress":"aa:bb:cc:dd:ee:ff"}'
  uictl api delete /v1/sites/{siteId}/networks/{networkId}

Classic API (requires --raw):
  uictl api get --raw /proxy/network/api/s/default/stat/device
  uictl api get --raw /proxy/network/api/s/default/rest/setting/ips
  uictl api put --raw /proxy/network/api/s/default/rest/device/{_id} --data '{"port_overrides":[...]}'`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAPI,
}

func init() {
	apiCmd.Flags().StringP("data", "d", "", "JSON request body")
	apiCmd.Flags().Bool("raw", false, "Bypass integration API prefix (for classic API endpoints)")
}

func runAPI(cmd *cobra.Command, args []string) error {
	method := strings.ToUpper(args[0])
	path := args[1]
	rawFlag, _ := cmd.Flags().GetBool("raw")

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	switch method {
	case "GET":
		var data []byte
		if rawFlag {
			data, err = client.GetRaw(path)
		} else {
			data, err = client.Get(path)
		}
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
		if rawFlag {
			switch method {
			case "POST":
				resp, err = client.PostRaw(path, body)
			case "PUT":
				resp, err = client.PutRaw(path, body)
			case "PATCH":
				resp, err = client.PutRaw(path, body) // classic API rarely uses PATCH
			}
		} else {
			switch method {
			case "POST":
				resp, err = client.Post(path, body)
			case "PUT":
				resp, err = client.Put(path, body)
			case "PATCH":
				resp, err = client.Patch(path, body)
			}
		}
		if err != nil {
			return err
		}
		if resp != nil {
			_, _ = fmt.Fprintln(os.Stdout, string(resp))
		}

	case "DELETE":
		if rawFlag {
			resp, doErr := client.DoRaw("DELETE", path, nil)
			if doErr != nil {
				return doErr
			}
			_ = resp.Body.Close()
		} else {
			if doErr := client.Delete(path); doErr != nil {
				return doErr
			}
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
