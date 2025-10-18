package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func NewVMDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(AccentColor).
		BorderLeft(true).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Copy()
	return d
}

func NewHostDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("212")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(AccentColor).
		BorderLeft(true).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("240"))
	return d
}
