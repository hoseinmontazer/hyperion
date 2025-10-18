package cmd

import (
	"context"
	"fmt"
	"hyperion/internal/vmware"

	"github.com/spf13/cobra"
)

var getVMsCmd = &cobra.Command{
	Use:   "vms",
	Short: "List all VMs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := vmware.LoadConfig("config/config.yaml")

		for _, host := range cfg.ESXi {
			client := vmware.ConnectESXI(ctx, host)
			fmt.Printf("\nVMs on host: %s\n", host.Host)
			vmware.ListVMs(ctx, client)
		}
	},
}
