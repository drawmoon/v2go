package internal

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
)

// Type representing the mode of the panel
type panelMode uint

// Type representing the focus type of the file panel
type filePanelFocusType uint

// Type representing the state of a process
type processState int

// Type representing the type of focused panel
type FocusPanelType int

type warnType int

type hotkeyType int

type channelMessageType int

const (
	globalType hotkeyType = iota
	normalType
	selectType
)

const (
	confirmDeleteItem warnType = iota
	confirmRenameItem
)

// Constants for panel with no focus
const (
	NoPanelFocus FocusPanelType = iota
	ProcessFocus
	SidebarFocus
	MetadataFocus
)

// Constants for file panel with no focus
const (
	noneFocus filePanelFocusType = iota
	secondFocus
	focus
)

// Constants for select mode or browser mode
const (
	selectMode panelMode = iota
	browserMode
)

// Constants for operation, success, cancel, failure
const (
	inOperation processState = iota
	successful
	cancel
	failure
)

const (
	sendWarnModal channelMessageType = iota
	sendMetadata
	sendProcess
)

// Modal
type commandLineModal struct {
	input textinput.Model
}

type warnModal struct {
	open     bool
	warnType warnType
	title    string
	content  string
}

type typingModal struct {
	location  string
	open      bool
	textInput textinput.Model
}

// File metadata
type Metadata struct {
	MetadataItems [][2]string
	RenderIndex   int
}

// Copied items
type copyItems struct {
	items []string
	cut   bool
}

/*PROCESS BAR internal TYPE START*/

// Model for an individual process
type process struct {
	name     string
	progress progress.Model
	state    processState
	total    int
	done     int
	doneTime time.Time
}

// Message for process bar
type channelMessage struct {
	messageId       string
	messageType     channelMessageType
	processNewState process
	warnModal       warnModal
	metadata        [][2]string
}

/*PROCESS BAR internal TYPE END*/

type editorFinishedMsg struct{ err error }
