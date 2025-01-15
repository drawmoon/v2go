package internal

import (
	variable "github.com/yorukot/superfile/src/config"
)

// Create new file panel
func (m *Model) createNewFilePanel() {
	if len(m.File.FilePanels) == m.File.MaxFilePanel {
		return
	}

	m.File.FilePanels = append(m.File.FilePanels, FilePanel{
		Location:        variable.UserHomeDir,
		SortOptions:     m.File.FilePanels[m.filePanelFocusIndex].SortOptions,
		PanelMode:       browserMode,
		FocusType:       secondFocus,
		DirectoryRecord: make(map[string]directoryRecord),
		SearchBar:       GetSearchModel(),
	})

	if m.File.Preview.Open {
		// File preview panel width same as file panel
		if Config.FilePreviewWidth == 0 {
			m.File.Preview.Width = (m.Context.WindowWidth - Config.SidebarWidth - (4 + (len(m.File.FilePanels))*2)) / (len(m.File.FilePanels) + 1)
		} else {
			m.File.Preview.Width = (m.Context.WindowWidth - Config.SidebarWidth) / Config.FilePreviewWidth
		}
	}

	m.File.FilePanels[m.filePanelFocusIndex].FocusType = noneFocus
	m.File.FilePanels[m.filePanelFocusIndex+1].FocusType = returnFocusType(m.Context.FocusPanel)
	m.File.Width = (m.Context.WindowWidth - Config.SidebarWidth - m.File.Preview.Width - (4 + (len(m.File.FilePanels)-1)*2)) / len(m.File.FilePanels)
	m.filePanelFocusIndex++

	m.File.MaxFilePanel = (m.Context.WindowWidth - Config.SidebarWidth - m.File.Preview.Width) / 20

	for i := range m.File.FilePanels {
		m.File.FilePanels[i].SearchBar.Width = m.File.Width - 4
	}
}

// Close current focus file panel
func (m *Model) closeFilePanel() {
	if len(m.File.FilePanels) == 1 {
		return
	}

	m.File.FilePanels = append(m.File.FilePanels[:m.filePanelFocusIndex], m.File.FilePanels[m.filePanelFocusIndex+1:]...)

	if m.File.Preview.Open {
		// File preview panel width same as file panel
		if Config.FilePreviewWidth == 0 {
			m.File.Preview.Width = (m.Context.WindowWidth - Config.SidebarWidth - (4 + (len(m.File.FilePanels))*2)) / (len(m.File.FilePanels) + 1)
		} else {

			m.File.Preview.Width = (m.Context.WindowWidth - Config.SidebarWidth) / Config.FilePreviewWidth
		}
	}

	if m.filePanelFocusIndex != 0 {
		m.filePanelFocusIndex--
	}

	m.File.Width = (m.Context.WindowWidth - Config.SidebarWidth - m.File.Preview.Width - (4 + (len(m.File.FilePanels)-1)*2)) / len(m.File.FilePanels)
	m.File.FilePanels[m.filePanelFocusIndex].FocusType = returnFocusType(m.Context.FocusPanel)

	m.File.MaxFilePanel = (m.Context.WindowWidth - Config.SidebarWidth - m.File.Preview.Width) / 20

	for i := range m.File.FilePanels {
		m.File.FilePanels[i].SearchBar.Width = m.File.Width - 4
	}
}

func (m *Model) toggleFilePreviewPanel() {
	m.File.Preview.Open = !m.File.Preview.Open
	m.File.Preview.Width = 0
	if m.File.Preview.Open {
		// File preview panel width same as file panel
		if Config.FilePreviewWidth == 0 {
			m.File.Preview.Width = (m.Context.WindowWidth - Config.SidebarWidth - (4 + (len(m.File.FilePanels))*2)) / (len(m.File.FilePanels) + 1)
		} else {
			m.File.Preview.Width = (m.Context.WindowWidth - Config.SidebarWidth) / Config.FilePreviewWidth
		}
	}

	m.File.Width = (m.Context.WindowWidth - Config.SidebarWidth - m.File.Preview.Width - (4 + (len(m.File.FilePanels)-1)*2)) / len(m.File.FilePanels)

	m.File.MaxFilePanel = (m.Context.WindowWidth - Config.SidebarWidth - m.File.Preview.Width) / 20

	for i := range m.File.FilePanels {
		m.File.FilePanels[i].SearchBar.Width = m.File.Width - 4
	}

}

// Focus on next file panel
func (m *Model) nextFilePanel() {
	m.File.FilePanels[m.filePanelFocusIndex].FocusType = noneFocus
	if m.filePanelFocusIndex == (len(m.File.FilePanels) - 1) {
		m.filePanelFocusIndex = 0
	} else {
		m.filePanelFocusIndex++
	}

	m.File.FilePanels[m.filePanelFocusIndex].FocusType = returnFocusType(m.Context.FocusPanel)
}

// Focus on previous file panel
func (m *Model) previousFilePanel() {
	m.File.FilePanels[m.filePanelFocusIndex].FocusType = noneFocus
	if m.filePanelFocusIndex == 0 {
		m.filePanelFocusIndex = (len(m.File.FilePanels) - 1)
	} else {
		m.filePanelFocusIndex--
	}

	m.File.FilePanels[m.filePanelFocusIndex].FocusType = returnFocusType(m.Context.FocusPanel)
}

// Focus on sidebar
func (m *Model) focusOnSideBar() {
	if Config.SidebarWidth == 0 {
		return
	}
	if m.Context.FocusPanel == SidebarFocus {
		m.Context.FocusPanel = NoPanelFocus
		m.File.FilePanels[m.filePanelFocusIndex].FocusType = focus
	} else {
		m.Context.FocusPanel = SidebarFocus
		m.File.FilePanels[m.filePanelFocusIndex].FocusType = secondFocus
	}
}

// Focus on processbar
func (m *Model) focusOnProcessBar() {
	if m.Context.FocusPanel == ProcessFocus {
		m.Context.FocusPanel = NoPanelFocus
		m.File.FilePanels[m.filePanelFocusIndex].FocusType = focus
	} else {
		m.Context.FocusPanel = ProcessFocus
		m.File.FilePanels[m.filePanelFocusIndex].FocusType = secondFocus
	}
}

// focus on metadata
func (m *Model) focusOnMetadata() {
	if m.Context.FocusPanel == MetadataFocus {
		m.Context.FocusPanel = NoPanelFocus
		m.File.FilePanels[m.filePanelFocusIndex].FocusType = focus
	} else {
		m.Context.FocusPanel = MetadataFocus
		m.File.FilePanels[m.filePanelFocusIndex].FocusType = secondFocus
	}
}
