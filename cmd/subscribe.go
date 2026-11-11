package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
)

var subscribeUpdatesModelKeys = map[string]struct{}{
	"nvr":                {},
	"camera":             {},
	"chime":              {},
	"light":              {},
	"viewer":             {},
	"speaker":            {},
	"bridge":             {},
	"sensor":             {},
	"siren":              {},
	"fob":                {},
	"relay":              {},
	"aiProcessor":        {},
	"aiPort":             {},
	"linkStation":        {},
	"event":              {},
	"externalArmProfile": {},
}

func init() {
	rootCmd.AddCommand(subscribeCmd)
}

var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Subscribe to Protect websocket streams",
	Long: `Print Protect websocket messages until SIGINT.

Examples:
  uictl subscribe updates --model-keys camera,event`,
}

func init() {
	subscribeCmd.AddCommand(subscribeUpdatesCmd)
	subscribeUpdatesCmd.Flags().String("model-keys", "", "Comma-separated model keys to subscribe to (required)")
}

var subscribeUpdatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "Subscribe to Protect model updates",
	Long: `Print Protect model create/update/remove frames until SIGINT.

--model-keys is required. Valid keys: nvr, camera, chime, light, viewer,
speaker, bridge, sensor, siren, fob, relay, aiProcessor, aiPort,
linkStation, event, externalArmProfile.

Examples:
  uictl subscribe updates --model-keys camera,event`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetString("model-keys")
		keys, err := parseSubscribeModelKeys(raw)
		if err != nil {
			return err
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		path := "/v1/subscribe/updates" + "?modelKeys=" + strings.Join(keys, ",")
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer cancel()
		err = client.Subscribe(ctx, path, func(msg []byte) error {
			_, werr := fmt.Fprintln(cmd.OutOrStdout(), string(msg))
			return werr
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	},
}

func parseSubscribeModelKeys(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("--model-keys is required")
	}
	var keys []string
	for _, part := range strings.Split(raw, ",") {
		key := strings.TrimSpace(part)
		if key == "" {
			continue
		}
		if _, ok := subscribeUpdatesModelKeys[key]; !ok {
			return nil, fmt.Errorf("invalid model key %q", key)
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("--model-keys is required")
	}
	return keys, nil
}
