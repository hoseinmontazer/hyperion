package tui

import (
	"context"
	"hyperion/internal/vmware"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Tab int

const (
	TabVM Tab = iota
	TabHosts
	TabUsage
	TabDetails
)

type VMItem struct {
	Host string
	Name string
	IP   string
}

func (i VMItem) Title() string       { return i.Name }
func (i VMItem) Description() string { return "IP: " + i.IP }
func (i VMItem) FilterValue() string { return i.Name }

type HostItem struct {
	Name     string
	Host     string
	IP       string
	CPUGHz   int
	RAMGB    int
	CPUUsage int
	MemUsage int
	VMCount  int
}

func (i HostItem) Title() string {
	return i.Name
}

func (i HostItem) Description() string {
	return "IP: " + i.IP
}
func (i HostItem) FilterValue() string { return i.Name }

type TickMsg struct{}

type Model struct {
	Ctx          context.Context
	Cfg          *vmware.Config
	Current      Tab
	VMLists      map[string]list.Model
	HostList     list.Model
	Details      viewport.Model
	Spinner      spinner.Model
	Progress     progress.Model
	Hosts        []vmware.HostInfo
	UsageBars    map[string][2]float64
	LastTick     time.Time
	Loading      bool
	WindowSize   tea.WindowSizeMsg
	SelectedHost string
}

// UI Styles
var (
	AccentColor  = lipgloss.Color("63")
	SuccessColor = lipgloss.Color("2")
	WarningColor = lipgloss.Color("3")
	ErrorColor   = lipgloss.Color("1")
	TextColor    = lipgloss.Color("230")
	SubtleColor  = lipgloss.Color("240")
)
