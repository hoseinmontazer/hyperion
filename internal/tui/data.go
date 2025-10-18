package tui

func (m *Model) SelectNextHost() {
	hosts := make([]string, 0, len(m.VMLists))
	for host := range m.VMLists {
		hosts = append(hosts, host)
	}

	if len(hosts) == 0 {
		return
	}

	currentIndex := -1
	for i, host := range hosts {
		if host == m.SelectedHost {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		m.SelectedHost = hosts[0]
	} else {
		m.SelectedHost = hosts[(currentIndex+1)%len(hosts)]
	}
}

func (m *Model) SelectPreviousHost() {
	hosts := make([]string, 0, len(m.VMLists))
	for host := range m.VMLists {
		hosts = append(hosts, host)
	}

	if len(hosts) == 0 {
		return
	}

	currentIndex := -1
	for i, host := range hosts {
		if host == m.SelectedHost {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		m.SelectedHost = hosts[0]
	} else {
		m.SelectedHost = hosts[(currentIndex-1+len(hosts))%len(hosts)]
	}
}
