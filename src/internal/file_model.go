package internal

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type FileModel struct {
	Width        int
	Renaming     bool
	MaxFilePanel int
	FilePanels   []FilePanel
	Preview      PreviewModel
}

// Panel representing a file
type FilePanel struct {
	Cursor             int
	Render             int
	FocusType          filePanelFocusType
	Location           string
	SortOptions        sortOptionsModel
	PanelMode          panelMode
	Selected           []string
	Element            []element
	DirectoryRecord    map[string]directoryRecord
	Rename             textinput.Model
	Renaming           bool
	SearchBar          textinput.Model
	LastTimeGetElement time.Time
}

// Sort options
type sortOptionsModel struct {
	width  int
	height int
	open   bool
	cursor int
	data   sortOptionsModelData
}

type sortOptionsModelData struct {
	options  []string
	selected int
	reversed bool
}

// Record for directory navigation
type directoryRecord struct {
	directoryCursor int
	directoryRender int
}

// Element within a file panel
type element struct {
	name      string
	location  string
	directory bool
	matchRate float64
	metaData  [][2]string
}

func NewFileModel(firstFilePanelDir string) FileModel {
	return FileModel{
		FilePanels: []FilePanel{
			{
				Render:   0,
				Cursor:   0,
				Location: firstFilePanelDir,
				SortOptions: sortOptionsModel{
					width:  20,
					height: 4,
					open:   false,
					cursor: Config.DefaultSortType,
					data: sortOptionsModelData{
						options:  []string{"Name", "Size", "Date Modified"},
						selected: Config.DefaultSortType,
						reversed: Config.SortOrderReversed,
					},
				},
				PanelMode:       browserMode,
				FocusType:       focus,
				DirectoryRecord: make(map[string]directoryRecord),
				SearchBar:       GetSearchModel(),
			},
		},
		Preview: NewPreviewModel(),
		Width:   10,
	}
}

func (m Model) renderFile() string {
	// file panel
	f := make([]string, 10)
	for i, filePanel := range m.File.FilePanels {
		// check if cursor or render out of range
		if filePanel.Cursor > len(filePanel.Element)-1 {
			filePanel.Cursor = 0
			filePanel.Render = 0
		}
		m.File.FilePanels[i] = filePanel

		f[i] += " " + filePanel.SearchBar.View() + "\n"

		for h := filePanel.Render; h < filePanel.Render+panelElementHeight(m.mainPanelHeight) && h < len(filePanel.Element); h++ {
			endl := "\n"
			if h == filePanel.Render+panelElementHeight(m.mainPanelHeight)-1 || h == len(filePanel.Element)-1 {
				endl = ""
			}

			var cursor Cursor
			// Check if the cursor needs to be displayed, if the user is using the search bar, the cursor is not displayed
			if h == filePanel.Cursor && !filePanel.SearchBar.Focused() {
				cursor = NewCursor()
			}

			isItemSelected := arrayContains(filePanel.Selected, filePanel.Element[h].location)
			if filePanel.Renaming && h == filePanel.Cursor {
				f[i] += filePanel.Rename.View() + endl
			} else {
				_, err := os.ReadDir(filePanel.Element[h].location)
				f[i] += cursor.View() + prettierName(filePanel.Element[h].name, m.File.Width-5, filePanel.Element[h].directory || (err == nil), isItemSelected, filePanelBGColor) + endl
			}
		}
		cursorPosition := strconv.Itoa(filePanel.Cursor + 1)
		totalElement := strconv.Itoa(len(filePanel.Element))

		m.File.Width = 45
		border := NewCard(m.File.Width, m.mainPanelHeight, "List", f[i], fmt.Sprintf("%s/%s", cursorPosition, totalElement), "", true)
		f[i] = border.View()
	}

	// file panel render together
	filePanelRender := ""
	for _, f := range f {
		filePanelRender = lipgloss.JoinHorizontal(lipgloss.Top, filePanelRender, f)
	}
	return filePanelRender
}
