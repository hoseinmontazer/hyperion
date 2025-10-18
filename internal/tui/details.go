package tui

import (
	"fmt"
	"hyperion/internal/vmware"
)

// func (m *Model) UpdateVMDetails(item VMItem) {
// 	details := fmt.Sprintf(
// 		"🏗️  Virtual Machine Details\n\n"+
// 			"Name: %s\n"+
// 			"Host: %s\n"+
// 			"Status: %s\n"+
// 			"Power: %s\n"+
// 			"CPU: %d vCPU\n"+
// 			"Memory: %d GB\n"+
// 			"Storage: %d GB",
// 		item.Name, item.Host, "Running", "Powered On", 4, 8, 100,
// 	)
// 	m.Details.SetContent(details)
// 	m.Current = TabDetails
// }

func (m *Model) UpdateVMDetails(item VMItem) {
	// Connect to the host
	var hCfg vmware.HostConfig
	for _, cfg := range m.Cfg.ESXi {
		if cfg.Host == item.Host {
			hCfg = cfg
			break
		}
	}
	client := vmware.ConnectESXI(m.Ctx, hCfg)

	// Get all VMs on that host
	vmInfos := vmware.GetVMInfos(m.Ctx, client)

	// Find the selected VM
	var selected *vmware.VMInfo
	for _, v := range vmInfos {
		if v.Name == item.Name {
			selected = &v
			break
		}
	}

	// If VM not found, show a message
	if selected == nil {
		m.Details.SetContent(fmt.Sprintf("VM %s not found on host %s", item.Name, item.Host))
		m.Current = TabDetails
		return
	}

	// Format VM details dynamically
	details := fmt.Sprintf(
		"🏗️  Virtual Machine Details\n\n"+
			"Name: %s\n"+
			"Host: %s\n"+
			"Power: %s\n"+
			"IP: %s\n"+
			"CPU: %d vCPU\n"+
			"Memory: %d MB\n"+
			"Guest OS: %s\n"+
			"Storage: %d GB\n"+
			"Networks: %v",
		selected.Name,
		item.Host,
		selected.Power,
		selected.IP,
		selected.CPU,
		selected.RAMMB,
		selected.GuestOS,
		selected.StorageGB,
		selected.Networks,
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
