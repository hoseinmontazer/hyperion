package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.WindowSize.Width < 60 || m.WindowSize.Height < 20 {
		return lipgloss.NewStyle().
			Foreground(ErrorColor).
			Padding(1, 2).
			Render("❌ Window too small! Please resize to at least 60x20")
	}

	// Header
	header := lipgloss.NewStyle().
		Foreground(TextColor).
		Background(AccentColor).
		Bold(true).
		Padding(0, 1).
		Width(m.WindowSize.Width).
		Render("🚀 Hyperion vSphere Dashboard")

	// Tabs
	tabs := RenderTabs(m.Current, m.WindowSize.Width)

	// Content
	var content string
	switch m.Current {
	case TabVM:
		if m.Loading {
			content = lipgloss.Place(
				m.WindowSize.Width-4, m.WindowSize.Height-10,
				lipgloss.Center, lipgloss.Center,
				fmt.Sprintf("%s Loading VMs...", m.Spinner.View()),
			)
		} else if m.SelectedHost != "" {
			if vmList, exists := m.VMLists[m.SelectedHost]; exists {
				// Show host selector
				hosts := make([]string, 0, len(m.VMLists))
				for host := range m.VMLists {
					hosts = append(hosts, host)
				}
				hostSelector := RenderHostSelector(hosts, m.SelectedHost, m.WindowSize.Width)
				content = hostSelector + vmList.View()
			} else {
				content = "No VMs found for selected host"
			}
		} else {
			content = "No hosts available"
		}
	case TabHosts:
		if m.Loading {
			content = lipgloss.Place(
				m.WindowSize.Width-4, m.WindowSize.Height-10,
				lipgloss.Center, lipgloss.Center,
				fmt.Sprintf("%s Loading hosts...", m.Spinner.View()),
			)
		} else {
			content = m.HostList.View()
		}
	case TabUsage:
		body := "\n"
		for _, h := range m.Hosts {
			bars := m.UsageBars[h.Name]
			body += lipgloss.NewStyle().Bold(true).Render(h.Name) + "\n"
			body += "  " + RenderUsageBar(bars[0], float64(h.CPUGHz*1000), 30, "CPU") + "\n"
			body += "  " + RenderUsageBar(bars[1], float64(h.RAMGB*1024), 30, "Memory") + "\n\n"
		}
		content = lipgloss.NewStyle().
			Width(m.WindowSize.Width - 4).
			Height(m.WindowSize.Height - 10).
			Render(body)
	case TabDetails:
		content = m.Details.View()
	}

	// Footer
	footer := RenderFooter(len(m.Cfg.ESXi), m.LastTick, m.Loading, m.WindowSize.Width, m.Current, m.SelectedHost)

	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)
}
