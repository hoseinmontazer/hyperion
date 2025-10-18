package cmd

import (
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources from ESXi / vCenter",
	Long:  `Get resources such as VMs, Hosts, or Usage`,
}

// Subcommands like get vms will be added here
func init() {
	getCmd.AddCommand(getVMsCmd)
}
