package internal

import tea "github.com/charmbracelet/bubbletea"

func containsKey(v string, a []string) string {
	for _, i := range a {
		if i == v {
			return v
		}
	}
	return ""
}

// mainKey handles most of key commands in the regular state of the application. For
// keys that performs actions in multiple panels, like going up or down,
// check the state of model m and handle properly.
func (m *Model) mainKey(msg string, cmd tea.Cmd) tea.Cmd {
	switch msg {

	// If move up Key is pressed, check the current state and executes
	case containsKey(msg, hotkeys.ListUp):
		if m.Context.FocusPanel == SidebarFocus {
			m.controlSideBarListUp(false)
		} else if m.Context.FocusPanel == ProcessFocus {
			m.controlProcessbarListUp(false)
		} else if m.Context.FocusPanel == MetadataFocus {
			m.controlMetadataListUp(false)
		} else if m.Context.FocusPanel == NoPanelFocus {
			m.controlFilePanelListUp(false)
			m.Metadata.RenderIndex = 0
			go func() {
				m.returnMetadata()
			}()
		}

		// If move down Key is pressed, check the current state and executes
	case containsKey(msg, hotkeys.ListDown):
		if m.Context.FocusPanel == SidebarFocus {
			m.controlSideBarListDown(false)
		} else if m.Context.FocusPanel == ProcessFocus {
			m.controlProcessbarListDown(false)
		} else if m.Context.FocusPanel == MetadataFocus {
			m.controlMetadataListDown(false)
		} else if m.Context.FocusPanel == NoPanelFocus {
			m.controlFilePanelListDown(false)
			m.Metadata.RenderIndex = 0
			go func() {
				m.returnMetadata()
			}()
		}

	case containsKey(msg, hotkeys.PageUp):
		m.controlFilePanelPgUp()

	case containsKey(msg, hotkeys.PageDown):
		m.controlFilePanelPgDown()

	case containsKey(msg, hotkeys.ChangePanelMode):
		m.changeFilePanelMode()

	case containsKey(msg, hotkeys.NextFilePanel):
		m.nextFilePanel()

	case containsKey(msg, hotkeys.PreviousFilePanel):
		m.previousFilePanel()

	case containsKey(msg, hotkeys.CloseFilePanel):
		m.closeFilePanel()

	case containsKey(msg, hotkeys.CreateNewFilePanel):
		m.createNewFilePanel()

	case containsKey(msg, hotkeys.ToggleFilePreviewPanel):
		m.toggleFilePreviewPanel()

	case containsKey(msg, hotkeys.FocusOnSidebar):
		m.focusOnSideBar()

	case containsKey(msg, hotkeys.FocusOnProcessBar):
		m.focusOnProcessBar()

	case containsKey(msg, hotkeys.FocusOnMetaData):
		m.focusOnMetadata()
		go func() {
			m.returnMetadata()
		}()

	case containsKey(msg, hotkeys.PasteItems):
		go func() {
			m.pasteItem()
		}()

	case containsKey(msg, hotkeys.FilePanelItemCreate):
		m.panelCreateNewFile()

	case containsKey(msg, hotkeys.ToggleDotFile):
		m.toggleDotFileController()

	case containsKey(msg, hotkeys.ExtractFile):
		go func() {
			m.extractFile()
		}()

	case containsKey(msg, hotkeys.CompressFile):
		go func() {
			m.compressFile()
		}()

	case containsKey(msg, hotkeys.OpenHelpMenu):
		m.openHelpMenu()

	case containsKey(msg, hotkeys.ToggleReverseSort):
		m.toggleReverseSort()

	case containsKey(msg, hotkeys.OpenCommandLine):
		m.openCommandLine()

	case containsKey(msg, hotkeys.OpenFileWithEditor):
		cmd = m.openFileWithEditor()

	case containsKey(msg, hotkeys.OpenCurrentDirectoryWithEditor):
		cmd = m.openDirectoryWithEditor()

	default:
		m.normalAndBrowserModeKey(msg)
	}

	return cmd
}

func (m *Model) normalAndBrowserModeKey(msg string) {
	// if not focus on the filepanel return
	if m.File.FilePanels[m.filePanelFocusIndex].FocusType != focus {
		if m.Context.FocusPanel == SidebarFocus && (msg == containsKey(msg, hotkeys.Confirm)) {
			m.sidebarSelectDirectory()
		}
		return
	}
	// Check if in the select mode and focusOn filepanel
	if m.File.FilePanels[m.filePanelFocusIndex].PanelMode == selectMode {
		switch msg {
		case containsKey(msg, hotkeys.Confirm):
			m.singleItemSelect()
		case containsKey(msg, hotkeys.FilePanelSelectModeItemsSelectUp):
			m.itemSelectUp(false)
		case containsKey(msg, hotkeys.FilePanelSelectModeItemsSelectDown):
			m.itemSelectDown(false)
		case containsKey(msg, hotkeys.DeleteItems):
			go func() {
				m.deleteItemWarn()
			}()
		case containsKey(msg, hotkeys.CopyItems):
			m.copyMultipleItem()
		case containsKey(msg, hotkeys.CutItems):
			m.cutMultipleItem()
		case containsKey(msg, hotkeys.FilePanelSelectAllItem):
			m.selectAllItem()
		}
		return
	}

	switch msg {
	case containsKey(msg, hotkeys.Confirm):
		m.enterPanel()
	case containsKey(msg, hotkeys.ParentDirectory):
		m.parentDirectory()
	case containsKey(msg, hotkeys.DeleteItems):
		go func() {
			m.deleteItemWarn()
		}()
	case containsKey(msg, hotkeys.CopyItems):
		m.copySingleItem()
	case containsKey(msg, hotkeys.CutItems):
		m.cutSingleItem()
	case containsKey(msg, hotkeys.FilePanelItemRename):
		m.panelItemRename()
	case containsKey(msg, hotkeys.SearchBar):
		m.searchBarFocus()
	case containsKey(msg, hotkeys.CopyPath):
		m.copyPath()
	case containsKey(msg, hotkeys.CopyPWD):
		m.copyPWD()
	}
}

// Check the hotkey to cancel operation or create file
func (m *Model) typingModalOpenKey(msg string) {
	switch msg {
	case containsKey(msg, hotkeys.CancelTyping):
		m.cancelTypingModal()
	case containsKey(msg, hotkeys.ConfirmTyping):
		m.createItem()
	}
}

func (m *Model) warnModalOpenKey(msg string) {
	switch msg {
	case containsKey(msg, hotkeys.Quit), containsKey(msg, hotkeys.CancelTyping):
		m.cancelWarnModal()
		if m.warnModal.warnType == confirmRenameItem {
			m.cancelRename()
		}
	case containsKey(msg, hotkeys.Confirm):
		m.warnModal.open = false
		switch m.warnModal.warnType {
		case confirmDeleteItem:
			panel := m.File.FilePanels[m.filePanelFocusIndex]
			if m.File.FilePanels[m.filePanelFocusIndex].PanelMode == selectMode {
				if !hasTrash || isExternalDiskPath(panel.Location) {
					go func() {
						m.completelyDeleteMultipleItems()
						m.File.FilePanels[m.filePanelFocusIndex].Selected = m.File.FilePanels[m.filePanelFocusIndex].Selected[:0]
					}()
				} else {
					go func() {
						m.deleteMultipleItems()
						m.File.FilePanels[m.filePanelFocusIndex].Selected = m.File.FilePanels[m.filePanelFocusIndex].Selected[:0]
					}()
				}
			} else {
				if !hasTrash || isExternalDiskPath(panel.Location) {
					go func() {
						m.completelyDeleteSingleItem()
					}()
				} else {
					go func() {
						m.deleteSingleItem()
					}()
				}

			}
		case confirmRenameItem:
			m.confirmRename()
		}
	}
}

// Handle key input to confirm or cancel and close quiting warn in SPF
func (m *Model) confirmToQuitSuperfile(msg string) bool {
	switch msg {
	case containsKey(msg, hotkeys.Quit), containsKey(msg, hotkeys.CancelTyping):
		m.cancelWarnModal()
		m.confirmToQuit = false
		return false
	case containsKey(msg, hotkeys.Confirm):
		return true
	default:
		return false
	}
}

func (m *Model) renamingKey(msg string) {
	switch msg {
	case containsKey(msg, hotkeys.CancelTyping):
		m.cancelRename()
	case containsKey(msg, hotkeys.ConfirmTyping):
		if m.IsRenamingConflicting() {
			m.warnModalForRenaming()
		} else {
			m.confirmRename()
		}
	}
}

// Check the key input and cancel or confirms the search
func (m *Model) focusOnSearchbarKey(msg string) {
	switch msg {
	case containsKey(msg, hotkeys.CancelTyping):
		m.cancelSearch()
	case containsKey(msg, hotkeys.ConfirmTyping):
		m.confirmSearch()
	}
}

// Check hotkey input in help menu. Possible actions are moving up, down
// and quiting the menu
func (m *Model) helpMenuKey(msg string) {
	switch msg {
	case containsKey(msg, hotkeys.ListUp):
		m.helpMenuListUp()
	case containsKey(msg, hotkeys.ListDown):
		m.helpMenuListDown()
	case containsKey(msg, hotkeys.Quit):
		m.quitHelpMenu()
	}
}

// Handle command line keys closing or entering command line
func (m *Model) commandLineKey(msg string) {
	switch msg {
	case containsKey(msg, hotkeys.CancelTyping):
		m.closeCommandLine()
	case containsKey(msg, hotkeys.ConfirmTyping):
		m.enterCommandLine()
	}
}
