package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Cancel typing modal e.g. create file or directory
func (m *Model) cancelTypingModal() {
	m.typingModal.textInput.Blur()
	m.typingModal.open = false
}

// Close warn modal
func (m *Model) cancelWarnModal() {
	m.warnModal.open = false
}

// Confirm to create file or directory
func (m *Model) createItem() {
	if !strings.HasSuffix(m.typingModal.textInput.Value(), "/") {
		path := filepath.Join(m.typingModal.location, m.typingModal.textInput.Value())
		path, _ = renameIfDuplicate(path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			outPutLog("Create item func (m *model)tion error", err)
		}
		f, err := os.Create(path)
		if err != nil {
			outPutLog("Create item func (m *model)tion create file error", err)
		}
		defer f.Close()
	} else {
		path := m.typingModal.location + "/" + m.typingModal.textInput.Value()
		err := os.MkdirAll(path, 0755)
		if err != nil {
			outPutLog("Create item func (m *model)tion create folder error", err)
		}
	}
	m.typingModal.open = false
	m.typingModal.textInput.Blur()
}

// Cancel rename file or directory
func (m *Model) cancelRename() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	panel.Rename.Blur()
	panel.Renaming = false
	m.File.Renaming = false
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Connfirm rename file or directory
func (m *Model) confirmRename() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	oldPath := panel.Element[panel.Cursor].location
	newPath := panel.Location + "/" + panel.Rename.Value()

	// Rename the file
	err := os.Rename(oldPath, newPath)
	if err != nil {
		outPutLog("Confirm func (m *model)tion rename error", err)
	}

	m.File.Renaming = false
	panel.Rename.Blur()
	panel.Renaming = false
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}
func (m *Model) toggleReverseSort() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	panel.SortOptions.data.reversed = !panel.SortOptions.data.reversed
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Cancel search, this will clear all searchbar input
func (m *Model) cancelSearch() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	panel.SearchBar.Blur()
	panel.SearchBar.SetValue("")
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Confirm search. This will exit the search bar and filter the files
func (m *Model) confirmSearch() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	panel.SearchBar.Blur()
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Help menu panel list up
func (m *Model) helpMenuListUp() {
	if m.Menu.Cursor > 1 {
		m.Menu.Cursor--
		if m.Menu.Cursor < m.Menu.RenderIndex {
			m.Menu.RenderIndex--
			if m.Menu.MenuItems[m.Menu.Cursor].SubTitle != "" {
				m.Menu.RenderIndex--
			}
		}
		if m.Menu.MenuItems[m.Menu.Cursor].SubTitle != "" {
			m.Menu.Cursor--
		}
	} else {
		m.Menu.Cursor = len(m.Menu.MenuItems) - 1
		m.Menu.RenderIndex = len(m.Menu.MenuItems) - m.Menu.Height
	}
}

// Help menu panel list down
func (m *Model) helpMenuListDown() {
	if len(m.Menu.MenuItems) == 0 {
		return
	}

	if m.Menu.Cursor < len(m.Menu.MenuItems)-1 {
		m.Menu.Cursor++
		if m.Menu.Cursor > m.Menu.RenderIndex+m.Menu.Height-1 {
			m.Menu.RenderIndex++
			if m.Menu.MenuItems[m.Menu.Cursor].SubTitle != "" {
				m.Menu.RenderIndex++
			}
		}
		if m.Menu.MenuItems[m.Menu.Cursor].SubTitle != "" {
			m.Menu.Cursor++
		}
	} else {
		m.Menu.Cursor = 1
		m.Menu.RenderIndex = 0
	}
}

// Toggle help menu
func (m *Model) openHelpMenu() {
	if m.Menu.Focus {
		m.Menu.Focus = false
		return
	}

	m.Menu.Focus = true
}

// Quit help menu
func (m *Model) quitHelpMenu() {
	m.Menu.Focus = false
}

// Command line
func (m *Model) openCommandLine() {
	m.firstTextInput = true
	bottomHeight--
	m.commandLine.input = generateCommandLineInputBox()
	m.commandLine.input.Width = m.Context.WindowWidth - 3
	m.commandLine.input.Focus()
}

func (m *Model) closeCommandLine() {
	bottomHeight++
	m.commandLine.input.SetValue("")
	m.commandLine.input.Blur()
}

// Exec a command line input inside the pointing file dir. Like opening the
// focused file in the text editor
func (m *Model) enterCommandLine() {
	focusPanelDir := ""
	for _, panel := range m.File.FilePanels {
		if panel.FocusType == focus {
			focusPanelDir = panel.Location
		}
	}
	cd := "cd " + focusPanelDir + " && "
	cmd := exec.Command("/bin/sh", "-c", cd+m.commandLine.input.Value())
	_, err := cmd.CombinedOutput()
	m.commandLine.input.SetValue("")
	if err != nil {
		return
	}
	m.commandLine.input.Blur()
	bottomHeight++
}
