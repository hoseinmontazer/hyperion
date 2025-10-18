package tui

import "fmt"

func (m *Model) UpdateVMDetails(item VMItem) {
	details := fmt.Sprintf(
		"🏗️  Virtual Machine Details\n\n"+
			"Name: %s\n"+
			"Host: %s\n"+
			"Status: %s\n"+
			"Power: %s\n"+
			"CPU: %d vCPU\n"+
			"Memory: %d GB\n"+
			"Storage: %d GB",
		item.Name, item.Host, "Running", "Powered On", 4, 8, 100,
	)
	m.Details.SetContent(details)
	m.Current = TabDetails
}

func (m *Model) UpdateHostDetails(item HostItem) {
	details := fmt.Sprintf(
		"🖥️  ESXi Host Details\n\n"+
			"Name: %s\n"+
			"Address: %s\n"+
			"CPU: %d GHz (%d%% usage)\n"+
			"Memory: %d GB (%d%% usage)\n"+
			"VMs: %d running\n"+
			"Connection: %s\n"+
			"Version: %s",
		item.Name, item.Host, item.CPUGHz, item.CPUUsage, item.RAMGB, item.MemUsage,
		item.VMCount, "Connected", "7.0.3",
	)
	m.Details.SetContent(details)
	m.Current = TabDetails
}
