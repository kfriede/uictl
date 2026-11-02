package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(protectFileCmd)
}

var protectFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage Protect device asset files",
	Long: `List and upload UniFi Protect device asset files.

Examples:
  uictl file list <file-type>
  uictl file upload <file-type> --file ./tone.mp3`,
}

func init() {
	protectFileCmd.AddCommand(protectFileListCmd)
	protectFileCmd.AddCommand(protectFileUploadCmd)
}

var protectFileListCmd = &cobra.Command{
	Use: "list <file-type>", Short: "List device asset files of a type", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var result []map[string]any
		if err := client.GetJSON(fmt.Sprintf("/v1/files/%s", args[0]), &result); err != nil {
			return err
		}
		return printAPIResult(toAnySlice(result))
	},
}

var protectFileUploadCmd = &cobra.Command{
	Use:   "upload <file-type>",
	Short: "Upload a device asset file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file is required")
		}
		if flagDryRun {
			printer.Status(fmt.Sprintf("[dry-run] Would upload %s as file type %s", filePath, args[0]))
			return nil
		}
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		client, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := client.PostFile(fmt.Sprintf("/v1/files/%s", args[0]), filepath.Base(filePath), f)
		if err != nil {
			return err
		}
		return mutationResult(resp, fmt.Sprintf("Uploaded %s", filepath.Base(filePath)))
	},
}

func init() {
	protectFileUploadCmd.Flags().String("file", "", "Path of the file to upload")
}
