package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	variable "github.com/yorukot/superfile/src/config"
)

// Change file panel mode (select mode or browser mode)
func (m *Model) changeFilePanelMode() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if panel.PanelMode == selectMode {
		panel.Selected = panel.Selected[:0]
		panel.PanelMode = browserMode
	} else if panel.PanelMode == browserMode {
		panel.PanelMode = selectMode
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Back to parent directory
func (m *Model) parentDirectory() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	panel.DirectoryRecord[panel.Location] = directoryRecord{
		directoryCursor: panel.Cursor,
		directoryRender: panel.Render,
	}
	fullPath := panel.Location
	parentDir := filepath.Dir(fullPath)
	panel.Location = parentDir
	directoryRecord, hasRecord := panel.DirectoryRecord[panel.Location]
	if hasRecord {
		panel.Cursor = directoryRecord.directoryCursor
		panel.Render = directoryRecord.directoryRender
	} else {
		panel.Cursor = 0
		panel.Render = 0
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Enter directory or open file with default application
func (m *Model) enterPanel() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]

	if len(panel.Element) == 0 {
		return
	}

	if panel.Element[panel.Cursor].directory {
		panel.DirectoryRecord[panel.Location] = directoryRecord{
			directoryCursor: panel.Cursor,
			directoryRender: panel.Render,
		}
		panel.Location = panel.Element[panel.Cursor].location
		directoryRecord, hasRecord := panel.DirectoryRecord[panel.Location]
		if hasRecord {
			panel.Cursor = directoryRecord.directoryCursor
			panel.Render = directoryRecord.directoryRender
		} else {
			panel.Cursor = 0
			panel.Render = 0
		}
		panel.SearchBar.SetValue("")
	} else if !panel.Element[panel.Cursor].directory {
		fileInfo, err := os.Lstat(panel.Element[panel.Cursor].location)
		if err != nil {
			outPutLog("err when getting file info", err)
			return
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			targetPath, err := filepath.EvalSymlinks(panel.Element[panel.Cursor].location)
			if err != nil {
				return
			}

			targetInfo, err := os.Lstat(targetPath)

			if err != nil {
				return
			}

			if targetInfo.IsDir() {
				m.File.FilePanels[m.filePanelFocusIndex].Location = targetPath
			}

			return
		}

		openCommand := "xdg-open"
		if runtime.GOOS == "darwin" {
			openCommand = "open"
		} else if runtime.GOOS == "windows" {

			dllpath := filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
			dllfile := "url.dll,FileProtocolHandler"

			cmd := exec.Command(dllpath, dllfile, panel.Element[panel.Cursor].location)
			err = cmd.Start()
			if err != nil {
				outPutLog(fmt.Sprintf("err when open file with %s", openCommand), err)
			}

			return
		}

		cmd := exec.Command(openCommand, panel.Element[panel.Cursor].location)
		err = cmd.Start()
		if err != nil {
			outPutLog(fmt.Sprintf("err when open file with %s", openCommand), err)
		}

	}

	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Switch to the directory where the sidebar cursor is located
func (m *Model) sidebarSelectDirectory() {
	m.Context.FocusPanel = NoPanelFocus
	panel := m.File.FilePanels[m.filePanelFocusIndex]

	panel.DirectoryRecord[panel.Location] = directoryRecord{
		directoryCursor: panel.Cursor,
		directoryRender: panel.Render,
	}

	panel.Location = m.Sidebar.Directories[m.Sidebar.Cursor].Location
	directoryRecord, hasRecord := panel.DirectoryRecord[panel.Location]
	if hasRecord {
		panel.Cursor = directoryRecord.directoryCursor
		panel.Render = directoryRecord.directoryRender
	} else {
		panel.Cursor = 0
		panel.Render = 0
	}
	panel.FocusType = focus
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Select all item in the file panel (only work on select mode)
func (m *Model) selectAllItem() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	for _, item := range panel.Element {
		panel.Selected = append(panel.Selected, item.location)
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Select the item where cursor located (only work on select mode)
func (m *Model) singleItemSelect() {
	panel := m.File.FilePanels[m.filePanelFocusIndex] // Access the current panel

	if len(panel.Element) > 0 && panel.Cursor >= 0 && panel.Cursor < len(panel.Element) {
		elementLocation := panel.Element[panel.Cursor].location

		if arrayContains(panel.Selected, elementLocation) {
			panel.Selected = removeElementByValue(panel.Selected, elementLocation)
		} else {
			panel.Selected = append(panel.Selected, elementLocation)
		}

		m.File.FilePanels[m.filePanelFocusIndex] = panel
	} else {
		outPutLog("No elements to select or cursor out of bounds.")
	}
}

// Toggle dotfile display or not
func (m *Model) toggleDotFileController() {
	newToggleDotFile := ""
	if m.toggleDotFile {
		newToggleDotFile = "false"
		m.toggleDotFile = false
	} else {
		newToggleDotFile = "true"
		m.toggleDotFile = true
	}
	m.updatedToggleDotFile = true
	err := os.WriteFile(variable.ToggleDotFile, []byte(newToggleDotFile), 0644)
	if err != nil {
		outPutLog("Pinned folder function updatedData superfile data error", err)
	}

}

// Focus on search bar
func (m *Model) searchBarFocus() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if panel.SearchBar.Focused() {
		panel.SearchBar.Blur()
	} else {
		panel.SearchBar.Focus()
	}

	// config search bar width
	panel.SearchBar.Width = m.File.Width - 4
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// ======================================== File panel controller ========================================

// Control file panel list up
func (m *Model) controlFilePanelListUp(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.File.FilePanels[m.filePanelFocusIndex]
		if len(panel.Element) == 0 {
			return
		}
		if panel.Cursor > 0 {
			panel.Cursor--
			if panel.Cursor < panel.Render {
				panel.Render--
			}
		} else {
			if len(panel.Element) > panelElementHeight(m.mainPanelHeight) {
				panel.Render = len(panel.Element) - panelElementHeight(m.mainPanelHeight)
				panel.Cursor = len(panel.Element) - 1
			} else {
				panel.Cursor = len(panel.Element) - 1
			}
		}

		m.File.FilePanels[m.filePanelFocusIndex] = panel
	}
}

// Control file panel list down
func (m *Model) controlFilePanelListDown(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.File.FilePanels[m.filePanelFocusIndex]
		if len(panel.Element) == 0 {
			return
		}
		if panel.Cursor < len(panel.Element)-1 {
			panel.Cursor++
			if panel.Cursor > panel.Render+panelElementHeight(m.mainPanelHeight)-1 {
				panel.Render++
			}
		} else {
			panel.Render = 0
			panel.Cursor = 0
		}
		m.File.FilePanels[m.filePanelFocusIndex] = panel
	}

}

func (m *Model) controlFilePanelPgUp() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	panlen := len(panel.Element)
	panHeight := panelElementHeight(m.mainPanelHeight)
	panCenter := panHeight / 2 // For making sure the cursor is at the center of the panel

	if panlen == 0 {
		return
	}

	if panHeight >= panlen {
		panel.Cursor = 0
	} else {
		if panel.Cursor-panHeight <= 0 {
			panel.Cursor = 0
			panel.Render = 0
		} else {
			panel.Cursor -= panHeight
			panel.Render = panel.Cursor - panCenter

			if panel.Render < 0 {
				panel.Render = 0
			}
		}
	}

	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

func (m *Model) controlFilePanelPgDown() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	panlen := len(panel.Element)
	panHeight := panelElementHeight(m.mainPanelHeight)
	panCenter := panHeight / 2 // For making sure the cursor is at the center of the panel

	if panlen == 0 {
		return
	}

	if panHeight >= panlen {
		panel.Cursor = panlen - 1
	} else {
		if panel.Cursor+panHeight >= panlen {
			panel.Cursor = panlen - 1
			panel.Render = panel.Cursor - panCenter
		} else {
			panel.Cursor += panHeight
			panel.Render = panel.Cursor - panCenter
		}
	}

	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Handles the action of selecting an item in the file panel upwards. (only work on select mode)
func (m *Model) itemSelectUp(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.File.FilePanels[m.filePanelFocusIndex]
		if panel.Cursor > 0 {
			panel.Cursor--
			if panel.Cursor < panel.Render {
				panel.Render--
			}
		} else {
			if len(panel.Element) > panelElementHeight(m.mainPanelHeight) {
				panel.Render = len(panel.Element) - panelElementHeight(m.mainPanelHeight)
				panel.Cursor = len(panel.Element) - 1
			} else {
				panel.Cursor = len(panel.Element) - 1
			}
		}
		selectItemIndex := panel.Cursor + 1
		if selectItemIndex > len(panel.Element)-1 {
			selectItemIndex = 0
		}
		if arrayContains(panel.Selected, panel.Element[selectItemIndex].location) {
			panel.Selected = removeElementByValue(panel.Selected, panel.Element[selectItemIndex].location)
		} else {
			panel.Selected = append(panel.Selected, panel.Element[selectItemIndex].location)
		}

		m.File.FilePanels[m.filePanelFocusIndex] = panel
	}
}

// Handles the action of selecting an item in the file panel downwards. (only work on select mode)
func (m *Model) itemSelectDown(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		panel := m.File.FilePanels[m.filePanelFocusIndex]
		if panel.Cursor < len(panel.Element)-1 {
			panel.Cursor++
			if panel.Cursor > panel.Render+panelElementHeight(m.mainPanelHeight)-1 {
				panel.Render++
			}
		} else {
			panel.Render = 0
			panel.Cursor = 0
		}
		selectItemIndex := panel.Cursor - 1
		if selectItemIndex < 0 {
			selectItemIndex = len(panel.Element) - 1
		}
		if arrayContains(panel.Selected, panel.Element[selectItemIndex].location) {
			panel.Selected = removeElementByValue(panel.Selected, panel.Element[selectItemIndex].location)
		} else {
			panel.Selected = append(panel.Selected, panel.Element[selectItemIndex].location)
		}

		m.File.FilePanels[m.filePanelFocusIndex] = panel
	}
}

// ======================================== Sidebar controller ========================================

// Yorukot: P.S God bless me, this sidebar controller code is really ugly...

// Control sidebar panel list up
func (m *Model) controlSideBarListUp(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.Sidebar.Cursor > 0 {
			m.Sidebar.Cursor--
		} else {
			m.Sidebar.Cursor = len(m.Sidebar.Directories) - 1
		}
		newDirectory := m.Sidebar.Directories[m.Sidebar.Cursor].Location

		for newDirectory == "Pinned+-*/=?" || newDirectory == "Disks+-*/=?" {
			m.Sidebar.Cursor--
			newDirectory = m.Sidebar.Directories[m.Sidebar.Cursor].Location
		}
		changeToPlus := false
		cursorRender := false
		for !cursorRender {
			totalHeight := 2
			for i := m.Sidebar.RenderIndex; i < len(m.Sidebar.Directories); i++ {
				if totalHeight >= m.mainPanelHeight {
					break
				}
				directory := m.Sidebar.Directories[i]

				if directory.Location == "Pinned+-*/=?" {
					totalHeight += 3
					continue
				}

				if directory.Location == "Disks+-*/=?" {
					if m.mainPanelHeight-totalHeight <= 2 {
						break
					}
					totalHeight += 3
					continue
				}

				totalHeight++
				if m.Sidebar.Cursor == i && m.Context.FocusPanel == SidebarFocus {
					cursorRender = true
				}
			}

			if changeToPlus {
				m.Sidebar.RenderIndex++
				continue
			}

			if !cursorRender {
				m.Sidebar.RenderIndex--
			}
			if m.Sidebar.RenderIndex < 0 {
				changeToPlus = true
				m.Sidebar.RenderIndex++
			}
		}

		if changeToPlus {
			m.Sidebar.RenderIndex--
		}
	}
}

// Control sidebar panel list down
func (m *Model) controlSideBarListDown(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		lenDirs := len(m.Sidebar.Directories)
		if m.Sidebar.Cursor < lenDirs-1 {
			m.Sidebar.Cursor++
		} else {
			m.Sidebar.Cursor = 0
		}

		newDirectory := m.Sidebar.Directories[m.Sidebar.Cursor].Location
		for newDirectory == "Pinned+-*/=?" || newDirectory == "Disks+-*/=?" {
			m.Sidebar.Cursor++
			if m.Sidebar.Cursor+1 > len(m.Sidebar.Directories) {
				m.Sidebar.Cursor = 0
			}
			newDirectory = m.Sidebar.Directories[m.Sidebar.Cursor].Location
		}
		cursorRender := false
		for !cursorRender {
			totalHeight := 2
			for i := m.Sidebar.RenderIndex; i < len(m.Sidebar.Directories); i++ {
				if totalHeight >= m.mainPanelHeight {
					break
				}

				directory := m.Sidebar.Directories[i]

				if directory.Location == "Pinned+-*/=?" {
					totalHeight += 3
					continue
				}

				if directory.Location == "Disks+-*/=?" {
					if m.mainPanelHeight-totalHeight <= 2 {
						break
					}
					totalHeight += 3
					continue
				}

				totalHeight++
				if m.Sidebar.Cursor == i && m.Context.FocusPanel == SidebarFocus {
					cursorRender = true
				}
			}

			if !cursorRender {
				m.Sidebar.RenderIndex++
			}
			if m.Sidebar.RenderIndex > m.Sidebar.Cursor {
				m.Sidebar.RenderIndex = 0
			}
		}
	}
}

// ======================================== Metadata controller ========================================

// Control metadata panel up
func (m *Model) controlMetadataListUp(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	if len(m.Metadata.MetadataItems) == 0 {
		return
	}

	for i := 0; i < runTime; i++ {
		if m.Metadata.RenderIndex > 0 {
			m.Metadata.RenderIndex--
		} else {
			m.Metadata.RenderIndex = len(m.Metadata.MetadataItems) - 1
		}
	}
}

// Control metadata panel down
func (m *Model) controlMetadataListDown(wheel bool) {
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.Metadata.RenderIndex < len(m.Metadata.MetadataItems)-1 {
			m.Metadata.RenderIndex++
		} else {
			m.Metadata.RenderIndex = 0
		}
	}
}

// ======================================== Processbar controller ========================================

// Control processbar panel list up
func (m *Model) controlProcessbarListUp(wheel bool) {
	if len(m.ProcessModel.ProcessList) == 0 {
		return
	}
	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.ProcessModel.Cursor > 0 {
			m.ProcessModel.Cursor--
			if m.ProcessModel.Cursor < m.ProcessModel.Render {
				m.ProcessModel.Render--
			}
		} else {
			if len(m.ProcessModel.ProcessList) <= 3 || (len(m.ProcessModel.ProcessList) <= 2 && bottomHeight < 14) {
				m.ProcessModel.Cursor = len(m.ProcessModel.ProcessList) - 1
			} else {
				m.ProcessModel.Render = len(m.ProcessModel.ProcessList) - 3
				m.ProcessModel.Cursor = len(m.ProcessModel.ProcessList) - 1
			}
		}
	}
}

// Control processbar panel list down
func (m *Model) controlProcessbarListDown(wheel bool) {
	if len(m.ProcessModel.ProcessList) == 0 {
		return
	}

	runTime := 1
	if wheel {
		runTime = wheelRunTime
	}

	for i := 0; i < runTime; i++ {
		if m.ProcessModel.Cursor < len(m.ProcessModel.ProcessList)-1 {
			m.ProcessModel.Cursor++
			if m.ProcessModel.Cursor > m.ProcessModel.Render+2 {
				m.ProcessModel.Render++
			}
		} else {
			m.ProcessModel.Render = 0
			m.ProcessModel.Cursor = 0
		}
	}
}
