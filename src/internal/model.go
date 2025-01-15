package internal

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	variable "github.com/yorukot/superfile/src/config"
	stringfunction "github.com/yorukot/superfile/src/pkg/string_function"
)

var (
	LastTimeCursorMove = [2]int{int(time.Now().UnixMicro()), 0}
	ListeningMessage   = true

	hasTrash = true

	theme   ThemeConfig
	Config  AppConfig
	hotkeys HotkeysConfig

	et *exiftool.Exiftool

	channel                             = make(chan channelMessage, 1000)
	progressBarLastRenderTime time.Time = time.Now()
)

type ViewModel interface {
	GetDisplayStrings() string
	GetOnFocus() bool
	GetKeybindings() interface{}
}

type ModelContext struct {
	WindowWidth  int
	WindowHeight int
	FocusPanel   FocusPanelType
}

type Model struct {
	Context              *ModelContext
	Title                string
	Sidebar              SidebarModel
	File                 FileModel
	ProcessModel         ProcessModel
	copyItems            copyItems
	typingModal          typingModal
	warnModal            warnModal
	Menu                 MenuModal
	Metadata             Metadata
	commandLine          commandLineModal
	confirmToQuit        bool
	firstTextInput       bool
	toggleDotFile        bool
	updatedToggleDotFile bool
	filePanelFocusIndex  int
	mainPanelHeight      int
}

// Initialize and return model with default configs
func NewModel(dir string, hasTrashCheck bool) Model {
	toggleDotFileBool, firstFilePanelDir := initialConfig(dir)
	hasTrash = hasTrashCheck

	ctx := &ModelContext{WindowWidth: 0, WindowHeight: 0, FocusPanel: NoPanelFocus}
	return Model{
		Context:             ctx,
		Title:               "SuperFile",
		filePanelFocusIndex: 0,
		ProcessModel: ProcessModel{
			Context: ctx,
			Process: make(map[string]process),
			Cursor:  0,
			Render:  0,
		},
		Sidebar:       NewSidebarModel(),
		File:          NewFileModel(firstFilePanelDir),
		Menu:          NewMenuModel(),
		toggleDotFile: toggleDotFileBool,
	}
}

// Init function to be called by Bubble tea framework, sets windows title,
// cursos blinking and starts message streamming channel
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle(m.Title),
		textinput.Blink, // Assuming textinput.Blink is a valid command
		listenForChannelMessage(channel),
	)
}

// Update function for bubble tea to provide internal communication to the
// application
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case channelMessage:
		m.handleChannelMessage(msg)
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg)
	case tea.MouseMsg:
		m, cmd = wheelMainAction(msg.String(), m, cmd)
	case tea.KeyMsg:
		m, cmd = m.handleKeyInput(msg, cmd)
	}

	m.updateFilePanelsState(msg, &cmd)
	m.Sidebar.Directories = getDirectories()

	// check if there already have listening message
	if !ListeningMessage {
		cmd = tea.Batch(cmd, listenForChannelMessage(channel))
	}

	m.getFilePanelItems()

	return m, tea.Batch(cmd)
}

// Handle message exchanging whithin the application
func (m *Model) handleChannelMessage(msg channelMessage) {
	switch msg.messageType {
	case sendWarnModal:
		m.warnModal = msg.warnModal
	case sendMetadata:
		m.Metadata.MetadataItems = msg.metadata
	default:
		if !arrayContains(m.ProcessModel.ProcessList, msg.messageId) {
			m.ProcessModel.ProcessList = append(m.ProcessModel.ProcessList, msg.messageId)
		}
		m.ProcessModel.Process[msg.messageId] = msg.processNewState
	}
}

// Adjust window size based on msg information
func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) {
	m.Context.WindowHeight = msg.Height
	m.Context.WindowWidth = msg.Width

	if m.File.Preview.Open {
		// File preview panel width same as file panel
		m.setFilePreviewWidth(msg.Width)
	}

	m.setFilePanelsSize(msg.Width)
	m.setFooterSize(msg.Height)
	m.setHelpMenuSize()

	if m.File.MaxFilePanel >= 10 {
		m.File.MaxFilePanel = 10
	}
}

// Set file preview panel Widht to width. Assure that
func (m *Model) setFilePreviewWidth(width int) {
	if Config.FilePreviewWidth == 0 {
		m.File.Preview.Width = (width - Config.SidebarWidth - (4 + (len(m.File.FilePanels))*2)) / (len(m.File.FilePanels) + 1)
	} else if Config.FilePreviewWidth > 10 || Config.FilePreviewWidth == 1 {
		log.Fatalln("Config file file_preview_width invalidation")
	} else {
		m.File.Preview.Width = (width - Config.SidebarWidth) / Config.FilePreviewWidth
	}
}

// Proper set panels size. Assure that panels do not overlap
func (m *Model) setFilePanelsSize(width int) {
	// set each file panel size and max file panel amount
	m.File.Width = (width - Config.SidebarWidth - m.File.Preview.Width - (4 + (len(m.File.FilePanels)-1)*2)) / len(m.File.FilePanels)
	m.File.MaxFilePanel = (width - Config.SidebarWidth - m.File.Preview.Width) / 20
	for i := range m.File.FilePanels {
		m.File.FilePanels[i].SearchBar.Width = m.File.Width - 4
	}
}

// Set footer size using height
func (m *Model) setFooterSize(height int) {
	if height < 30 {
		bottomHeight = 10
	} else if height < 35 {
		bottomHeight = 11
	} else if height < 40 {
		bottomHeight = 12
	} else if height < 45 {
		bottomHeight = 13
	} else {
		bottomHeight = 14
	}

	if m.commandLine.input.Focused() {
		bottomHeight--
	}

	m.mainPanelHeight = height - bottomHeight + 1
}

// Set help menu size
func (m *Model) setHelpMenuSize() {
	m.Menu.Height = m.Context.WindowHeight - 2
	m.Menu.Width = m.Context.WindowWidth - 2

	if m.Context.WindowHeight > 35 {
		m.Menu.Height = 30
	}

	if m.Context.WindowWidth > 95 {
		m.Menu.Width = 90
	}
}

// Identify the current state of the application m and properly handle the
// msg keybind pressed
func (m Model) handleKeyInput(msg tea.KeyMsg, cmd tea.Cmd) (Model, tea.Cmd) {
	if m.typingModal.open {
		m.typingModalOpenKey(msg.String())
	} else if m.warnModal.open {
		m.warnModalOpenKey(msg.String())
		// If renaming a object
	} else if m.File.Renaming {
		m.renamingKey(msg.String())
		// If search bar is open
	} else if m.File.FilePanels[m.filePanelFocusIndex].SearchBar.Focused() {
		m.focusOnSearchbarKey(msg.String())
		// If help menu is open
	} else if m.Menu.Focus {
		m.helpMenuKey(msg.String())
		// If command line input is send
	} else if m.commandLine.input.Focused() {
		m.commandLineKey(msg.String())
		// If asking to confirm quiting
	} else if m.confirmToQuit {
		quit := m.confirmToQuitSuperfile(msg.String())
		if quit {
			m.quitSuperfile()
			return m, tea.Quit
		}
		// If quiting input pressed, check if has any runing process and displays a
		// warn. Otherwise just quits application
	} else if msg.String() == containsKey(msg.String(), hotkeys.Quit) {
		if m.hasRunningProcesses() {
			m.warnModalForQuit()
			return m, cmd
		}

		m.quitSuperfile()
		return m, tea.Quit
	} else {
		// Handles general kinds of inputs in the regular state of the application
		cmd = m.mainKey(msg.String(), cmd)
	}
	return m, cmd
}

// Update the file panel state. Change name of renamed files, filter out files
// in search, update typingb bar, etc
func (m *Model) updateFilePanelsState(msg tea.Msg, cmd *tea.Cmd) {
	focusPanel := &m.File.FilePanels[m.filePanelFocusIndex]
	if m.firstTextInput {
		m.firstTextInput = false
	} else if m.File.Renaming {
		focusPanel.Rename, *cmd = focusPanel.Rename.Update(msg)
	} else if focusPanel.SearchBar.Focused() {
		focusPanel.SearchBar, *cmd = focusPanel.SearchBar.Update(msg)
		for _, hotkey := range hotkeys.SearchBar {
			if hotkey == focusPanel.SearchBar.Value() {
				focusPanel.SearchBar.SetValue("")
				break
			}
		}
	} else if m.commandLine.input.Focused() {
		m.commandLine.input, *cmd = m.commandLine.input.Update(msg)
	} else if m.typingModal.open {
		m.typingModal.textInput, *cmd = m.typingModal.textInput.Update(msg)
	}

	if focusPanel.Cursor < 0 {
		focusPanel.Cursor = 0
	}
}

// Check if there's any processes running in background
func (m *Model) hasRunningProcesses() bool {
	for _, data := range m.ProcessModel.Process {
		if data.state == inOperation && data.done != data.total {
			return true
		}
	}
	return false
}

// Triggers a warn for confirm quiting
func (m *Model) warnModalForQuit() {
	m.confirmToQuit = true
	m.warnModal.title = "Confirm to quit superfile"
	m.warnModal.content = "You still have files being processed. Are you sure you want to exit?"
}

// Implement View function for bubble tea model to handle visualization.
func (m Model) View() string {
	// check is the terminal size enough
	if m.Context.WindowHeight < minimumHeight || m.Context.WindowWidth < minimumWidth {
		return m.terminalSizeWarnRender()
	}
	if m.File.Width < 18 {
		return m.terminalSizeWarnAfterFirstRender()
	}

	var view string

	sidebarView := m.renderSidebar()
	fileView := m.renderFile()
	previewView := m.renderPreview()
	mainView := lipgloss.JoinHorizontal(0, sidebarView, fileView, previewView)

	processView := m.ProcessModel.View()
	metadataView := m.metadataRender()
	clipboardView := m.clipboardRender()
	footerView := lipgloss.JoinHorizontal(0, processView, metadataView, clipboardView)

	view = lipgloss.JoinVertical(0, mainView, footerView)
	if m.commandLine.input.Focused() {
		commandLineView := m.commandLineInputBoxRender()
		view = lipgloss.JoinVertical(0, view, commandLineView)
	}

	// check if need pop up modal
	if m.Menu.Focus {
		return m.prepareView(view)
	}

	if m.confirmToQuit {
		warnModal := m.warnModalRender()
		overlayX := m.Context.WindowWidth/2 - modalWidth/2
		overlayY := m.Context.WindowHeight/2 - modalHeight/2
		return stringfunction.PlaceOverlay(overlayX, overlayY, warnModal, view)
	}

	return view
}

// Returns a tea.cmd responsible for listening messages from msg channel
func listenForChannelMessage(msg chan channelMessage) tea.Cmd {
	return func() tea.Msg {
		for {
			m := <-msg
			if m.messageType != sendProcess {
				ListeningMessage = false
				return m
			}
			if time.Since(progressBarLastRenderTime).Seconds() > 2 || m.processNewState.state == successful || m.processNewState.done < 2 {
				ListeningMessage = false
				progressBarLastRenderTime = time.Now()
				return m
			}
		}
	}
}

// Render and update file panel items. Check for changes and updates in files and
// folders in the current directory.
func (m *Model) getFilePanelItems() {
	focusPanel := m.File.FilePanels[m.filePanelFocusIndex]
	for i, filePanel := range m.File.FilePanels {
		var fileElement []element
		nowTime := time.Now()
		// Check last time each element was updated, if less then 3 seconds ignore
		if filePanel.FocusType == noneFocus && nowTime.Sub(filePanel.LastTimeGetElement) < 3*time.Second {
			if !m.updatedToggleDotFile {
				continue
			}
		}

		focusPanelReRender := false

		if len(focusPanel.Element) > 0 {
			if filepath.Dir(focusPanel.Element[0].location) != focusPanel.Location {
				focusPanelReRender = true
			}
		} else {
			focusPanelReRender = true
		}

		reRenderTime := int(float64(len(filePanel.Element)) / 100)

		if filePanel.FocusType != noneFocus && nowTime.Sub(filePanel.LastTimeGetElement) < time.Duration(reRenderTime)*time.Second && !focusPanelReRender {
			continue
		}

		// Get file names based on search bar filter
		if filePanel.SearchBar.Value() != "" {
			fileElement = returnFolderElementBySearchString(filePanel.Location, m.toggleDotFile, filePanel.SearchBar.Value())
		} else {
			fileElement = returnFolderElement(filePanel.Location, m.toggleDotFile, filePanel.SortOptions.data)
		}
		// Update file panel list
		filePanel.Element = fileElement
		m.File.FilePanels[i].Element = fileElement
		m.File.FilePanels[i].LastTimeGetElement = nowTime
	}

	m.updatedToggleDotFile = false
}

// Close superfile application. Cd into the curent dir if CdOnQuit on and save
// the path in state direcotory
func (m Model) quitSuperfile() {
	// close exiftool session
	if Config.Metadata {
		et.Close()
	}
	// cd on quit
	currentDir := m.File.FilePanels[m.filePanelFocusIndex].Location
	variable.LastDir = currentDir

	if Config.CdOnQuit {
		// escape single quote
		currentDir = strings.ReplaceAll(currentDir, "'", "'\\''")
		os.WriteFile(variable.AppStateDir+"/lastdir", []byte("cd '"+currentDir+"'"), 0755)
	}
}
