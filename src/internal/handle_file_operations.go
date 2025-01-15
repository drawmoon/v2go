package internal

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lithammer/shortuuid"
	"github.com/yorukot/superfile/src/config/icon"
)

// Create a file in the currently focus file panel
func (m *Model) panelCreateNewFile() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	ti := textinput.New()
	ti.Cursor.Style = modalCursorStyle
	ti.Cursor.TextStyle = modalStyle
	ti.TextStyle = modalStyle
	ti.Cursor.Blink = true
	ti.Placeholder = "Add \"/\" represent Transcend folder at the end"
	ti.PlaceholderStyle = modalStyle
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = modalWidth - 10

	m.typingModal.location = panel.Location
	m.typingModal.open = true
	m.typingModal.textInput = ti
	m.firstTextInput = true

	m.File.FilePanels[m.filePanelFocusIndex] = panel

}

func (m *Model) IsRenamingConflicting() bool {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	oldPath := panel.Element[panel.Cursor].location
	newPath := panel.Location + "/" + panel.Rename.Value()

	if oldPath == newPath {
		return false
	}

	_, err := os.Stat(newPath)
	return err == nil
}

func (m *Model) warnModalForRenaming() {
	id := shortuuid.New()
	message := channelMessage{
		messageId:   id,
		messageType: sendWarnModal,
	}

	message.warnModal = warnModal{
		open:     true,
		title:    "There is already a file or directory with that name",
		content:  "This operation will override the existing file",
		warnType: confirmRenameItem,
	}
	channel <- message
}

// Rename file where the cusror is located
func (m *Model) panelItemRename() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if len(panel.Element) == 0 {
		return
	}
	ti := textinput.New()
	ti.Cursor.Style = filePanelCursorStyle
	ti.Cursor.TextStyle = filePanelStyle
	ti.Prompt = filePanelCursorStyle.Render(icon.Cursor + " ")
	ti.TextStyle = modalStyle
	ti.Cursor.Blink = true
	ti.Placeholder = "New name"
	ti.PlaceholderStyle = modalStyle
	ti.SetValue(panel.Element[panel.Cursor].name)
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = m.File.Width - 4

	m.File.Renaming = true
	panel.Renaming = true
	m.firstTextInput = true
	panel.Rename = ti
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

func (m *Model) deleteItemWarn() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if !((panel.PanelMode == selectMode && len(panel.Selected) != 0) || (panel.PanelMode == browserMode)) {
		return
	}
	id := shortuuid.New()
	message := channelMessage{
		messageId:   id,
		messageType: sendWarnModal,
	}

	if !hasTrash || isExternalDiskPath(panel.Location) {
		message.warnModal = warnModal{
			open:     true,
			title:    "Are you sure you want to completely delete",
			content:  "This operation cannot be undone and your data will be completely lost.",
			warnType: confirmDeleteItem,
		}
		channel <- message
		return
	} else {
		message.warnModal = warnModal{
			open:     true,
			title:    "Are you sure you want to move this to trash can",
			content:  "This operation will move file or directory to trash can.",
			warnType: confirmDeleteItem,
		}
		channel <- message
		return
	}
}

// Move file or directory to the trash can
func (m *Model) deleteSingleItem() {
	id := shortuuid.New()
	panel := m.File.FilePanels[m.filePanelFocusIndex]

	if len(panel.Element) == 0 {
		return
	}

	prog := progress.New(generateGradientColor())
	prog.PercentageStyle = footerStyle

	newProcess := process{
		name:     icon.Delete + icon.Space + panel.Element[panel.Cursor].name,
		progress: prog,
		state:    inOperation,
		total:    1,
		done:     0,
	}
	m.ProcessModel.Process[id] = newProcess

	message := channelMessage{
		messageId:       id,
		messageType:     sendProcess,
		processNewState: newProcess,
	}

	channel <- message
	err := trashMacOrLinux(panel.Element[panel.Cursor].location)

	if err != nil {
		p := m.ProcessModel.Process[id]
		p.state = failure
		message.processNewState = p
		channel <- message
	} else {
		p := m.ProcessModel.Process[id]
		p.done = 1
		p.state = successful
		p.doneTime = time.Now()
		message.processNewState = p
		channel <- message
	}
	if len(panel.Element) == 0 {
		panel.Cursor = 0
	} else {
		if panel.Cursor >= len(panel.Element) {
			panel.Cursor = len(panel.Element) - 1
		}
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Move file or directory to the trash can
func (m Model) deleteMultipleItems() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if len(panel.Selected) != 0 {
		id := shortuuid.New()
		prog := progress.New(generateGradientColor())
		prog.PercentageStyle = footerStyle

		newProcess := process{
			name:     icon.Delete + icon.Space + filepath.Base(panel.Selected[0]),
			progress: prog,
			state:    inOperation,
			total:    len(panel.Selected),
			done:     0,
		}

		m.ProcessModel.Process[id] = newProcess

		message := channelMessage{
			messageId:       id,
			messageType:     sendProcess,
			processNewState: newProcess,
		}

		channel <- message

		for _, filePath := range panel.Selected {

			p := m.ProcessModel.Process[id]
			p.name = icon.Delete + icon.Space + filepath.Base(filePath)
			p.done++
			p.state = inOperation
			if len(channel) < 5 {
				message.processNewState = p
				channel <- message
			}
			err := trashMacOrLinux(filePath)

			if err != nil {
				p.state = failure
				message.processNewState = p
				channel <- message
				outPutLog("Delete multiple item function error", err)
				m.ProcessModel.Process[id] = p
				break
			} else {
				if p.done == p.total {
					p.state = successful
					message.processNewState = p
					channel <- message
				}
				m.ProcessModel.Process[id] = p
			}
		}
	}

	if panel.Cursor >= len(panel.Element)-len(panel.Selected)-1 {
		panel.Cursor = len(panel.Element) - len(panel.Selected) - 1
		if panel.Cursor < 0 {
			panel.Cursor = 0
		}
	}
	panel.Selected = panel.Selected[:0]
}

// Completely delete file or folder (Not move to the trash can)
func (m *Model) completelyDeleteSingleItem() {
	id := shortuuid.New()
	panel := m.File.FilePanels[m.filePanelFocusIndex]

	if len(panel.Element) == 0 {
		return
	}

	prog := progress.New(generateGradientColor())
	prog.PercentageStyle = footerStyle

	newProcess := process{
		name:     "󰆴 " + panel.Element[panel.Cursor].name,
		progress: prog,
		state:    inOperation,
		total:    1,
		done:     0,
	}
	m.ProcessModel.Process[id] = newProcess

	message := channelMessage{
		messageId:       id,
		messageType:     sendProcess,
		processNewState: newProcess,
	}

	channel <- message

	err := os.RemoveAll(panel.Element[panel.Cursor].location)
	if err != nil {
		outPutLog("Completely delete single item function remove file error", err)
	}

	if err != nil {
		p := m.ProcessModel.Process[id]
		p.state = failure
		message.processNewState = p
		channel <- message
	} else {
		p := m.ProcessModel.Process[id]
		p.done = 1
		p.state = successful
		p.doneTime = time.Now()
		message.processNewState = p
		channel <- message
	}
	if len(panel.Element) == 0 {
		panel.Cursor = 0
	} else {
		if panel.Cursor >= len(panel.Element) {
			panel.Cursor = len(panel.Element) - 1
		}
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Completely delete all file or folder from clipboard (Not move to the trash can)
func (m Model) completelyDeleteMultipleItems() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if len(panel.Selected) != 0 {
		id := shortuuid.New()
		prog := progress.New(generateGradientColor())
		prog.PercentageStyle = footerStyle

		newProcess := process{
			name:     icon.Delete + icon.Space + filepath.Base(panel.Selected[0]),
			progress: prog,
			state:    inOperation,
			total:    len(panel.Selected),
			done:     0,
		}

		m.ProcessModel.Process[id] = newProcess

		message := channelMessage{
			messageId:       id,
			messageType:     sendProcess,
			processNewState: newProcess,
		}

		channel <- message
		for _, filePath := range panel.Selected {

			p := m.ProcessModel.Process[id]
			p.name = icon.Delete + icon.Space + filepath.Base(filePath)
			p.done++
			p.state = inOperation
			if len(channel) < 5 {
				message.processNewState = p
				channel <- message
			}
			err := os.RemoveAll(filePath)
			if err != nil {
				outPutLog("Completely delete multiple item function remove file error", err)
			}

			if err != nil {
				p.state = failure
				message.processNewState = p
				channel <- message
				outPutLog("Completely delete multiple item function error", err)
				m.ProcessModel.Process[id] = p
				break
			} else {
				if p.done == p.total {
					p.state = successful
					message.processNewState = p
					channel <- message
				}
				m.ProcessModel.Process[id] = p
			}
		}
	}

	if panel.Cursor >= len(panel.Element)-len(panel.Selected)-1 {
		panel.Cursor = len(panel.Element) - len(panel.Selected) - 1
		if panel.Cursor < 0 {
			panel.Cursor = 0
		}
	}
	panel.Selected = panel.Selected[:0]
}

// Copy directory or file
func (m *Model) copySingleItem() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	m.copyItems.cut = false
	m.copyItems.items = m.copyItems.items[:0]
	if len(panel.Element) == 0 {
		return
	}
	m.copyItems.items = append(m.copyItems.items, panel.Element[panel.Cursor].location)
	fileInfo, err := os.Stat(panel.Element[panel.Cursor].location)
	if os.IsNotExist(err) {
		m.copyItems.items = m.copyItems.items[:0]
		return
	}
	if err != nil {
		outPutLog("Copy single item get file state error", panel.Element[panel.Cursor].location, err)
	}

	if !fileInfo.IsDir() && float64(fileInfo.Size())/(1024*1024) < 250 {
		fileContent, err := os.ReadFile(panel.Element[panel.Cursor].location)

		if err != nil {
			outPutLog("Copy single item read file error", panel.Element[panel.Cursor].location, err)
		}

		if err := clipboard.WriteAll(string(fileContent)); err != nil {
			outPutLog("Copy single item write file error", panel.Element[panel.Cursor].location, err)
		}
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Copy all selected file or directory to the clipboard
func (m *Model) copyMultipleItem() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	m.copyItems.cut = false
	m.copyItems.items = m.copyItems.items[:0]
	if len(panel.Selected) == 0 {
		return
	}
	m.copyItems.items = panel.Selected
	fileInfo, err := os.Stat(panel.Selected[0])
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		outPutLog("Copy multiple item function get file state error", panel.Selected[0], err)
	}

	if !fileInfo.IsDir() && float64(fileInfo.Size())/(1024*1024) < 250 {
		fileContent, err := os.ReadFile(panel.Selected[0])

		if err != nil {
			outPutLog("Copy multiple item function read file error", err)
		}

		if err := clipboard.WriteAll(string(fileContent)); err != nil {
			outPutLog("Copy multiple item function write file to clipboard error", err)
		}
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Cut directory or file
func (m *Model) cutSingleItem() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	m.copyItems.cut = true
	m.copyItems.items = m.copyItems.items[:0]
	if len(panel.Element) == 0 {
		return
	}
	m.copyItems.items = append(m.copyItems.items, panel.Element[panel.Cursor].location)
	fileInfo, err := os.Stat(panel.Element[panel.Cursor].location)
	if os.IsNotExist(err) {
		m.copyItems.items = m.copyItems.items[:0]
		return
	}
	if err != nil {
		outPutLog("Cut single item get file state error", panel.Element[panel.Cursor].location, err)
	}

	if !fileInfo.IsDir() && float64(fileInfo.Size())/(1024*1024) < 250 {
		fileContent, err := os.ReadFile(panel.Element[panel.Cursor].location)

		if err != nil {
			outPutLog("Cut single item read file error", panel.Element[panel.Cursor].location, err)
		}

		if err := clipboard.WriteAll(string(fileContent)); err != nil {
			outPutLog("Cut single item write file error", panel.Element[panel.Cursor].location, err)
		}
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Cut all selected file or directory to the clipboard
func (m *Model) cutMultipleItem() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	m.copyItems.cut = true
	m.copyItems.items = m.copyItems.items[:0]
	if len(panel.Selected) == 0 {
		return
	}
	m.copyItems.items = panel.Selected
	fileInfo, err := os.Stat(panel.Selected[0])
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		outPutLog("Copy multiple item function get file state error", panel.Selected[0], err)
	}

	if !fileInfo.IsDir() && float64(fileInfo.Size())/(1024*1024) < 250 {
		fileContent, err := os.ReadFile(panel.Selected[0])

		if err != nil {
			outPutLog("Copy multiple item function read file error", err)
		}

		if err := clipboard.WriteAll(string(fileContent)); err != nil {
			outPutLog("Copy multiple item function write file to clipboard error", err)
		}
	}
	m.File.FilePanels[m.filePanelFocusIndex] = panel
}

// Paste all clipboard items
func (m Model) pasteItem() {
	id := shortuuid.New()
	panel := m.File.FilePanels[m.filePanelFocusIndex]

	if len(m.copyItems.items) == 0 {
		return
	}

	totalFiles := 0

	for _, folderPath := range m.copyItems.items {
		count, err := countFiles(folderPath)
		if err != nil {
			continue
		}
		totalFiles += count
	}

	prog := progress.New(generateGradientColor())
	prog.PercentageStyle = footerStyle

	prefixIcon := icon.Copy + icon.Space
	if m.copyItems.cut {
		prefixIcon = icon.Cut + icon.Space
	}

	newProcess := process{
		name:     prefixIcon + filepath.Base(m.copyItems.items[0]),
		progress: prog,
		state:    inOperation,
		total:    totalFiles,
		done:     0,
	}

	m.ProcessModel.Process[id] = newProcess

	message := channelMessage{
		messageId:       id,
		messageType:     sendProcess,
		processNewState: newProcess,
	}

	channel <- message

	p := m.ProcessModel.Process[id]
	for _, filePath := range m.copyItems.items {
		var err error
		if m.copyItems.cut && !isExternalDiskPath(filePath) {
			p.name = icon.Cut + icon.Space + filepath.Base(filePath)
		} else {
			if m.copyItems.cut {
				p.name = icon.Cut + icon.Space + filepath.Base(filePath)
			}
			p.name = icon.Copy + icon.Space + filepath.Base(filePath)
		}

		errMessage := "cut item error"
		if m.copyItems.cut && !isExternalDiskPath(filePath) {
			err = moveElement(filePath, filepath.Join(panel.Location, path.Base(filePath)))
		} else {
			newModel, err := pasteDir(filePath, filepath.Join(panel.Location, path.Base(filePath)), id, m)
			if err != nil {
				errMessage = "paste item error"
			}
			m = newModel
			if m.copyItems.cut {
				os.RemoveAll(filePath)
			}
		}
		p = m.ProcessModel.Process[id]
		if err != nil {
			p.state = failure
			message.processNewState = p
			channel <- message
			outPutLog(errMessage, err)
			m.ProcessModel.Process[id] = p
			break
		}
	}

	p.state = successful
	p.done = totalFiles
	p.doneTime = time.Now()
	message.processNewState = p
	channel <- message

	m.ProcessModel.Process[id] = p
	m.copyItems.cut = false
}

// Extrach compress file
func (m Model) extractFile() {
	var err error
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	ext := strings.ToLower(filepath.Ext(panel.Element[panel.Cursor].location))
	outputDir := fileNameWithoutExtension(panel.Element[panel.Cursor].location)
	outputDir, err = renameIfDuplicate(outputDir)

	if err != nil {
		outPutLog("Error extract file when create new directory", err)
	}

	switch ext {
	case ".zip":
		os.MkdirAll(outputDir, 0755)
		err = unzip(panel.Element[panel.Cursor].location, outputDir)
		if err != nil {
			outPutLog("Error extract file,", err)
		}
	default:
		os.MkdirAll(outputDir, 0755)
		err = extractCompressFile(panel.Element[panel.Cursor].location, outputDir)
		if err != nil {
			outPutLog("Error extract file,", err)
		}
	}
}

// Compress file or directory
func (m Model) compressFile() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	fileName := filepath.Base(panel.Element[panel.Cursor].location)

	zipName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".zip"
	zipName, err := renameIfDuplicate(zipName)

	if err != nil {
		outPutLog("Error compress file when rename duplicate", err)
	}

	zipSource(panel.Element[panel.Cursor].location, filepath.Join(filepath.Dir(panel.Element[panel.Cursor].location), zipName))
}

// Open file with default editor
func (m Model) openFileWithEditor() tea.Cmd {
	panel := m.File.FilePanels[m.filePanelFocusIndex]

	editor := Config.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "nano"
	}
	c := exec.Command(editor, panel.Element[panel.Cursor].location)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

// Open directory with default editor
func (m Model) openDirectoryWithEditor() tea.Cmd {
	editor := Config.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "nano"
	}
	c := exec.Command(editor, m.File.FilePanels[m.filePanelFocusIndex].Location)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

// Copy file path
func (m Model) copyPath() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if err := clipboard.WriteAll(panel.Element[panel.Cursor].location); err != nil {
		outPutLog("Copy path error", panel.Element[panel.Cursor].location, err)
	}
}

func (m Model) copyPWD() {
	panel := m.File.FilePanels[m.filePanelFocusIndex]
	if err := clipboard.WriteAll(panel.Location); err != nil {
		outPutLog("Copy present working directory error", panel.Location, err)
	}
}
