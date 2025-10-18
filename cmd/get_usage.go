package cmd

import (
	"context"
	"fmt"
	"hyperion/internal/vmware"

	"github.com/spf13/cobra"
)

var getUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show cluster resource usage",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := vmware.LoadConfig("config/config.yaml")

		for _, host := range cfg.ESXi {
			client := vmware.ConnectESXI(ctx, host)
			fmt.Printf("\nResource usage for host: %s\n", host.Host)
			vmware.ShowHostUsage(ctx, client)
		}
	},
}

func init() {
	getCmd.AddCommand(getUsageCmd)
}
