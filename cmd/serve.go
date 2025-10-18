package cmd

import (
	"context"
	"fmt"
	"hyperion/internal/tui"
	"hyperion/internal/vmware"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch Hyperion TUI Dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := vmware.LoadConfig("config/config.yaml")

		m := tui.NewModel(ctx, cfg)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if err := p.Start(); err != nil {
			fmt.Println("Error running TUI:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
