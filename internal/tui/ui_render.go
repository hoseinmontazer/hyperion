package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func RenderTabs(current Tab, width int) string {
	names := []string{"🎯 Virtual Machines", "🖥️ Hosts", "📊 Usage", "📋 Details"}
	out := ""

	for i, n := range names {
		style := lipgloss.NewStyle().
			Padding(0, 2).
			Bold(true)

		if Tab(i) == current {
			style = style.
				Background(AccentColor).
				Foreground(TextColor)
		} else {
			style = style.
				Foreground(SubtleColor).
				Background(lipgloss.Color("236"))
		}
		out += style.Render(n)
	}

	// Add refresh indicator
	refreshStyle := lipgloss.NewStyle().
		Foreground(SubtleColor).
		Padding(0, 2)
	out += refreshStyle.Render("🔄 Space to refresh")

	return lipgloss.NewStyle().Width(width).Render(out) + "\n"
}

func RenderHostSelector(hosts []string, selected string, width int) string {
	out := "Select Host: "
	for _, host := range hosts {
		style := lipgloss.NewStyle().Padding(0, 1)
		if host == selected {
			style = style.
				Background(AccentColor).
				Foreground(TextColor).
				Bold(true)
		} else {
			style = style.
				Foreground(SubtleColor).
				Background(lipgloss.Color("236"))
		}
		// Shorten host for display
		displayHost := host
		if len(host) > 20 {
			displayHost = host[:17] + "..."
		}
		out += style.Render(displayHost)
	}
	return lipgloss.NewStyle().Width(width).Render(out) + "\n"
}

func RenderUsageBar(current, max float64, width int, label string) string {
	pct := current / max
	if max == 0 {
		pct = 0
	}

	filled := int(float64(width) * pct)
	empty := width - filled

	color := SuccessColor
	if pct > 0.5 && pct <= 0.8 {
		color = WarningColor
	} else if pct > 0.8 {
		color = ErrorColor
	}

	bar := lipgloss.NewStyle().
		Foreground(color).
		Render("▰"+strings.Repeat("▰", filled)) +
		lipgloss.NewStyle().
			Foreground(SubtleColor).
			Render(strings.Repeat("▱", empty))

	percentage := fmt.Sprintf("%.1f%%", pct*100)

	return fmt.Sprintf("%-12s %s %s", label, bar, percentage)
}

func RenderFooter(numHosts int, lastRefresh time.Time, loading bool, width int, currentTab Tab, selectedHost string) string {
	status := "✅ Connected"
	if loading {
		status = "🔄 Refreshing..."
	}

	var controls string
	switch currentTab {
	case TabVM:
		controls = fmt.Sprintf("←/→: Switch Host | d: Details | Current: %s", selectedHost)
	case TabHosts:
		controls = "↑/↓: Navigate | d: Details"
	case TabUsage:
		controls = "Viewing resource usage"
	case TabDetails:
		controls = "Viewing details - tab to return"
	}

	content := fmt.Sprintf(
		"%s to %d host(s) | %s | Last refresh: %s | tab: Switch | space: Refresh | q: Quit",
		status, numHosts, controls, lastRefresh.Format("15:04:05"),
	)

	return lipgloss.NewStyle().
		Foreground(TextColor).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(width).
		Render(content)
}
