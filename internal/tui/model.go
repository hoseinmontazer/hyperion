package tui

import (
	"context"
	"fmt"
	"hyperion/internal/vmware"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func NewModel(ctx context.Context, cfg *vmware.Config) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(AccentColor)

	prog := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF85"))

	m := Model{
		Ctx:          ctx,
		Cfg:          cfg,
		Current:      TabVM,
		VMLists:      make(map[string]list.Model),
		UsageBars:    make(map[string][2]float64),
		LastTick:     time.Now(),
		Loading:      true,
		Spinner:      sp,
		Progress:     prog,
		WindowSize:   tea.WindowSizeMsg{Width: 80, Height: 24},
		SelectedHost: "",
	}

	// Initialize details viewport
	m.Details = viewport.New(60, 10)
	m.Details.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(0, 1)

	m.RefreshData()
	return m
}

func (m *Model) RefreshData() {
	m.Loading = true

	hostItems := []list.Item{}
	usage := make(map[string][2]float64)
	hosts := []vmware.HostInfo{}

	// Clear existing VM lists
	m.VMLists = make(map[string]list.Model)

	for _, hCfg := range m.Cfg.ESXi {
		client := vmware.ConnectESXI(m.Ctx, hCfg)
		vms := vmware.GetVMNames(m.Ctx, client)

		// Create VM list for this host
		vmItems := []list.Item{}
		for _, vm := range vms {
			vmItems = append(vmItems, VMItem{
				Host: hCfg.Host,
				Name: vm,
			})
		}

		// Create list for this host's VMs
		vmList := list.New(vmItems, NewVMDelegate(), m.WindowSize.Width-10, m.WindowSize.Height-10)
		vmList.SetShowTitle(true)
		vmList.SetShowHelp(false)
		vmList.SetShowStatusBar(false)
		vmList.SetFilteringEnabled(true)
		vmList.Title = fmt.Sprintf("VMs on %s", hCfg.Host)
		vmList.Styles.Title = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true).
			Padding(0, 1)

		m.VMLists[hCfg.Host] = vmList

		// Get host info
		h := vmware.GetHostInfo(m.Ctx, client)
		hosts = append(hosts, h)
		hostItems = append(hostItems, HostItem{
			Name:     h.Name,
			Host:     hCfg.Host,
			CPUGHz:   int(h.CPUGHz),
			RAMGB:    int(h.RAMGB),
			CPUUsage: int(h.CPUUsage),
			MemUsage: int(h.MemUsage),
			VMCount:  len(vms),
		})
		usage[h.Name] = [2]float64{float64(h.CPUUsage), float64(h.MemUsage)}
	}

	// Set first host as selected if none selected
	if m.SelectedHost == "" && len(m.Cfg.ESXi) > 0 {
		m.SelectedHost = m.Cfg.ESXi[0].Host
	}

	// Update host list
	m.HostList = list.New(hostItems, NewHostDelegate(), m.WindowSize.Width-10, m.WindowSize.Height-10)
	m.HostList.SetShowTitle(false)
	m.HostList.SetShowHelp(false)
	m.HostList.SetShowStatusBar(false)
	m.HostList.SetFilteringEnabled(true)
	m.HostList.Title = "ESXi Hosts"
	m.HostList.Styles.Title = lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true).
		Padding(0, 1)

	m.UsageBars = usage
	m.Hosts = hosts
	m.LastTick = time.Now()
	m.Loading = false
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		Tick(),
		m.Spinner.Tick,
	)
}

func Tick() tea.Cmd {
	return tea.Tick(time.Second*10, func(time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowSize = msg
		// Update all VM lists
		for host, vmList := range m.VMLists {
			vmList.SetSize(msg.Width-4, msg.Height-10)
			m.VMLists[host] = vmList
		}
		m.HostList.SetSize(msg.Width-4, msg.Height-10)
		m.Details.Width = msg.Width - 4
		m.Details.Height = 8
	case TickMsg:
		m.RefreshData()
		cmds = append(cmds, Tick())
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			m.Current = (m.Current + 1) % 4
			m.Details.GotoTop()
		case "shift+tab":
			m.Current = (m.Current - 1 + 4) % 4
			m.Details.GotoTop()
		case " ":
			m.RefreshData()
		case "d":
			if m.Current == TabVM && m.SelectedHost != "" {
				if vmList, exists := m.VMLists[m.SelectedHost]; exists && len(vmList.Items()) > 0 {
					if item, ok := vmList.SelectedItem().(VMItem); ok {
						m.UpdateVMDetails(item)
					}
				}
			} else if m.Current == TabHosts && len(m.HostList.Items()) > 0 {
				if item, ok := m.HostList.SelectedItem().(HostItem); ok {
					m.UpdateHostDetails(item)
					// Also select this host for VM view
					m.SelectedHost = item.Host
				}
			}
		case "left", "h":
			if m.Current == TabVM {
				m.SelectPreviousHost()
			}
		case "right", "l":
			if m.Current == TabVM {
				m.SelectNextHost()
			}
		}
	}

	if m.Loading {
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update the current component
	switch m.Current {
	case TabVM:
		if m.SelectedHost != "" {
			if vmList, exists := m.VMLists[m.SelectedHost]; exists {
				var cmd tea.Cmd
				updatedList, cmd := vmList.Update(msg)
				m.VMLists[m.SelectedHost] = updatedList
				cmds = append(cmds, cmd)
			}
		}
	case TabHosts:
		var cmd tea.Cmd
		m.HostList, cmd = m.HostList.Update(msg)
		cmds = append(cmds, cmd)
	case TabDetails:
		var cmd tea.Cmd
		m.Details, cmd = m.Details.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
