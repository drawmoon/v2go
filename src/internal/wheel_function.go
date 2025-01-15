package internal

import tea "github.com/charmbracelet/bubbletea"

func wheelMainAction(msg string, m Model, cmd tea.Cmd) (Model, tea.Cmd) {
	switch msg {

	case "wheel up":
		if m.Context.FocusPanel == SidebarFocus {
			m.controlSideBarListUp(true)
		} else if m.Context.FocusPanel == ProcessFocus {
			m.controlProcessbarListUp(true)
		} else if m.Context.FocusPanel == MetadataFocus {
			m.controlMetadataListUp(true)
		} else if m.Context.FocusPanel == NoPanelFocus {
			m.controlFilePanelListUp(true)
			m.Metadata.RenderIndex = 0
			go func() {
				m.returnMetadata()
			}()
		}

	case "wheel down":
		if m.Context.FocusPanel == SidebarFocus {
			m.controlSideBarListDown(true)
		} else if m.Context.FocusPanel == ProcessFocus {
			m.controlProcessbarListDown(true)
		} else if m.Context.FocusPanel == MetadataFocus {
			m.controlMetadataListDown(true)
		} else if m.Context.FocusPanel == NoPanelFocus {
			m.controlFilePanelListDown(true)
			m.Metadata.RenderIndex = 0
			go func() {
				m.returnMetadata()
			}()
		}
	}
	return m, cmd
}
