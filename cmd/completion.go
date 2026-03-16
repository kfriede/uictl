package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// registerDynamicCompletions sets up dynamic completion functions
// that query the UniFi API for resource names/IDs.
func registerDynamicCompletions() {
	// Device ID completion for commands that take a device-id argument
	deviceIDCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeResource("devices", "id", "name", toComplete)
	}

	// Client ID completion
	clientIDCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeResource("clients", "id", "name", toComplete)
	}

	// Network ID completion
	networkIDCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeResource("networks", "id", "name", toComplete)
	}

	// Register on device subcommands that take an ID
	for _, sub := range []*cobra.Command{deviceGetCmd, deviceRemoveCmd, deviceActionCmd, deviceStatsCmd} {
		sub.ValidArgsFunction = deviceIDCompletion
	}

	// Register on client subcommands
	for _, sub := range []*cobra.Command{clientGetCmd, clientAuthorizeCmd, clientUnauthorizeCmd} {
		sub.ValidArgsFunction = clientIDCompletion
	}

	// Register on network subcommands
	for _, sub := range []*cobra.Command{networkGetCmd, networkUpdateCmd, networkDeleteCmd, networkRefsCmd} {
		sub.ValidArgsFunction = networkIDCompletion
	}

	// Site flag completion
	_ = rootCmd.RegisterFlagCompletionFunc("site", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := newAPIClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		data, err := client.GetAllPages("/v1/sites")
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, item := range data {
			name, _ := item["name"].(string)
			if name != "" && strings.HasPrefix(strings.ToLower(name), strings.ToLower(toComplete)) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	})

	// Output format completion
	_ = rootCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "csv", "ndjson"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// completeResource queries the API for a resource list and returns
// matching completions in the format "id\tdescription".
func completeResource(resource, idField, nameField, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := newAPIClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	siteId := cfg.Site
	if siteId == "" {
		siteId = "default"
	}

	path := fmt.Sprintf("/v1/sites/%s/%s", siteId, resource)
	data, err := client.GetAllPages(path)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, item := range data {
		id, _ := item[idField].(string)
		name, _ := item[nameField].(string)
		if id == "" {
			continue
		}

		// Match against both ID and name
		lower := strings.ToLower(toComplete)
		if toComplete == "" ||
			strings.HasPrefix(strings.ToLower(id), lower) ||
			strings.HasPrefix(strings.ToLower(name), lower) {
			if name != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
			} else {
				completions = append(completions, id)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	// Delay registration until after all commands are added
	cobra.OnInitialize(registerDynamicCompletions)
}
