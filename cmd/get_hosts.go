package cmd

import (
	"context"
	"fmt"
	"hyperion/internal/vmware"

	"github.com/spf13/cobra"
)

var getHostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "List all ESXi hosts and resources",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := vmware.LoadConfig()

		for _, host := range cfg.ESXi {
			client := vmware.ConnectESXI(ctx, host)
			fmt.Printf("\nHost info: %s\n", host.Host)
			vmware.ListHostResources(ctx, client)
		}
	},
}

func init() {
	getCmd.AddCommand(getHostsCmd)
}
