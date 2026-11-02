package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(protectSubscribeCmd)
}

var protectSubscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Subscribe to Protect websocket streams",
	Long: `Print Protect websocket messages until SIGINT.

Examples:
  uictl subscribe devices
  uictl subscribe events`,
}

func init() {
	protectSubscribeCmd.AddCommand(protectSubscribeDevicesCmd)
	protectSubscribeCmd.AddCommand(protectSubscribeEventsCmd)
}

var protectSubscribeDevicesCmd = &cobra.Command{
	Use: "devices", Short: "Subscribe to Protect device updates", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)
		go func() {
			select {
			case <-sigCh:
				cancel()
			case <-ctx.Done():
			}
		}()
		err = client.Subscribe(ctx, "/v1/subscribe/devices", func(msg []byte) error {
			_, werr := fmt.Fprintln(cmd.OutOrStdout(), string(msg))
			return werr
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	},
}

var protectSubscribeEventsCmd = &cobra.Command{
	Use: "events", Short: "Subscribe to Protect events", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)
		go func() {
			select {
			case <-sigCh:
				cancel()
			case <-ctx.Done():
			}
		}()
		err = client.Subscribe(ctx, "/v1/subscribe/events", func(msg []byte) error {
			_, werr := fmt.Fprintln(cmd.OutOrStdout(), string(msg))
			return werr
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	},
}
