package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(countryCmd)
	rootCmd.AddCommand(dpiCmd)
	rootCmd.AddCommand(deviceTagCmd)
	rootCmd.AddCommand(radiusCmd)
	rootCmd.AddCommand(wanCmd)
	rootCmd.AddCommand(vpnCmd)
}

// --- Countries ---

var countryCmd = &cobra.Command{
	Use:   "country",
	Short: "List country codes",
}

func init() {
	countryCmd.AddCommand(countryListCmd)
}

var countryListCmd = &cobra.Command{
	Use: "list", Short: "List available countries", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages("/v1/countries")
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

// --- DPI ---

var dpiCmd = &cobra.Command{
	Use:   "dpi",
	Short: "View DPI applications and categories",
}

func init() {
	dpiCmd.AddCommand(dpiAppCmd)
	dpiCmd.AddCommand(dpiCatCmd)
}

var dpiAppCmd = &cobra.Command{
	Use: "app", Short: "List DPI applications", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages("/v1/dpi/applications")
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var dpiCatCmd = &cobra.Command{
	Use: "category", Short: "List DPI categories", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages("/v1/dpi/categories")
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

// --- Device Tags ---

var deviceTagCmd = &cobra.Command{
	Use:   "device-tag",
	Short: "View device tags",
}

func init() {
	deviceTagCmd.AddCommand(deviceTagListCmd)
}

var deviceTagListCmd = &cobra.Command{
	Use: "list", Short: "List device tags", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/device-tags", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

// --- RADIUS Profiles ---

var radiusCmd = &cobra.Command{
	Use:   "radius",
	Short: "View RADIUS profiles",
}

func init() {
	radiusCmd.AddCommand(radiusListCmd)
}

var radiusListCmd = &cobra.Command{
	Use: "list", Short: "List RADIUS profiles", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/radius/profiles", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

// --- WAN Interfaces ---

var wanCmd = &cobra.Command{
	Use:   "wan",
	Short: "View WAN interfaces",
}

func init() {
	wanCmd.AddCommand(wanListCmd)
}

var wanListCmd = &cobra.Command{
	Use: "list", Short: "List WAN interfaces", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/wans", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

// --- VPN ---

var vpnCmd = &cobra.Command{
	Use:   "vpn",
	Short: "View VPN servers and tunnels",
}

func init() {
	vpnCmd.AddCommand(vpnServerCmd)
	vpnCmd.AddCommand(vpnTunnelCmd)
}

var vpnServerCmd = &cobra.Command{
	Use: "server", Short: "List VPN servers", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/vpn/servers", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}

var vpnTunnelCmd = &cobra.Command{
	Use: "tunnel", Short: "List site-to-site VPN tunnels", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		siteId, err := requireSite()
		if err != nil {
			return err
		}
		data, err := client.GetAllPages(fmt.Sprintf("/v1/sites/%s/vpn/site-to-site-tunnels", siteId))
		if err != nil {
			return err
		}
		return printAPIResult(toAnySlice(data))
	},
}
