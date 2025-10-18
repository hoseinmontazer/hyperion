package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hyperion",
	Short: "Hyperion - ESXi / vCenter monitoring CLI",
	Long:  `Hyperion is a CLI tool to monitor ESXi / vCenter hosts and VMs`,
}

// Execute runs the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getHostsCmd)
	getCmd.AddCommand(getUsageCmd)
}
