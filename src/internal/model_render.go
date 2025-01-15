package internal

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/yorukot/superfile/src/config/icon"
)

type ProcessModel struct {
	Context     *ModelContext
	Render      int
	Cursor      int
	ProcessList []string
	Process     map[string]process
}

func (m ProcessModel) View() string {
	// save process in the array
	var processes []process
	for _, p := range m.Process {
		processes = append(processes, p)
	}

	// sort by the process
	sort.Slice(processes, func(i, j int) bool {
		doneI := (processes[i].state == successful)
		doneJ := (processes[j].state == successful)

		// sort by done or not
		if doneI != doneJ {
			return !doneI
		}

		// if both not done
		if !doneI {
			completionI := float64(processes[i].done) / float64(processes[i].total)
			completionJ := float64(processes[j].done) / float64(processes[j].total)
			return completionI < completionJ // Those who finish first will be ranked later.
		}

		// if both done sort by the doneTime
		return processes[j].doneTime.Before(processes[i].doneTime)
	})

	view := ""
	renderTimes := 0

	for i := m.Render; i < len(processes); i++ {
		if bottomHeight < 14 && renderTimes == 2 {
			break
		}
		if renderTimes == 3 {
			break
		}
		process := processes[i]
		process.progress.Width = footerWidth(m.Context.WindowWidth) - 3
		symbol := ""
		cursor := ""
		if i == m.Cursor {
			cursor = footerCursorStyle.Render("┃ ")
		} else {
			cursor = footerCursorStyle.Render("  ")
		}
		switch process.state {
		case failure:
			symbol = processErrorStyle.Render(icon.Warn)
		case successful:
			symbol = processSuccessfulStyle.Render(icon.Done)
		case inOperation:
			symbol = processInOperationStyle.Render(icon.InOperation)
		case cancel:
			symbol = processCancelStyle.Render(icon.Error)
		}

		view += cursor + footerStyle.Render(truncateText(process.name, footerWidth(m.Context.WindowWidth)-7, "...")+" ") + symbol + "\n"
		if renderTimes == 2 {
			view += cursor + process.progress.ViewAs(float64(process.done)/float64(process.total)) + ""
		} else if bottomHeight < 14 && renderTimes == 1 {
			view += cursor + process.progress.ViewAs(float64(process.done)/float64(process.total))
		} else {
			view += cursor + process.progress.ViewAs(float64(process.done)/float64(process.total)) + "\n\n"
		}
		renderTimes++
	}

	if len(processes) == 0 {
		view += "\n " + icon.Error + "  No processes running"
	}
	courseNumber := 0
	if len(m.ProcessList) == 0 {
		courseNumber = 0
	} else {
		courseNumber = m.Cursor + 1
	}

	borderWidth := footerWidth(m.Context.WindowWidth)
	borderHight := bottomElementHeight(bottomHeight)

	border := NewCard(borderWidth, borderHight, "Processes", view, fmt.Sprintf("%s/%s", strconv.Itoa(courseNumber), strconv.Itoa(len(m.ProcessList))), "p", m.Context.FocusPanel == ProcessFocus)
	view = border.View()
	return view
}

func (m Model) metadataRender() string {
	view := ""
	if len(m.Metadata.MetadataItems) == 0 && len(m.File.FilePanels[m.filePanelFocusIndex].Element) > 0 && !m.File.Renaming {
		m.Metadata.MetadataItems = append(m.Metadata.MetadataItems, [2]string{"", ""})
		m.Metadata.MetadataItems = append(m.Metadata.MetadataItems, [2]string{" " + icon.InOperation + "  Loading metadata...", ""})
		go func() {
			m.returnMetadata()
		}()
	}
	maxKeyLength := 0
	sort.Slice(m.Metadata.MetadataItems, func(i, j int) bool {
		comparisonFields := []string{"FileName", "FileSize", "FolderName", "FolderSize", "FileModifyDate", "FileAccessDate"}

		for _, field := range comparisonFields {
			if m.Metadata.MetadataItems[i][0] == field {
				return true
			} else if m.Metadata.MetadataItems[j][0] == field {
				return false
			}
		}

		// Default comparison
		return m.Metadata.MetadataItems[i][0] < m.Metadata.MetadataItems[j][0]
	})
	for _, data := range m.Metadata.MetadataItems {
		if len(data[0]) > maxKeyLength {
			maxKeyLength = len(data[0])
		}
	}

	sprintfLength := maxKeyLength + 1
	valueLength := footerWidth(m.Context.WindowWidth) - maxKeyLength - 2
	if valueLength < footerWidth(m.Context.WindowWidth)/2 {
		valueLength = footerWidth(m.Context.WindowWidth)/2 - 2
		sprintfLength = valueLength
	}

	for i := m.Metadata.RenderIndex; i < bottomElementHeight(bottomHeight)+m.Metadata.RenderIndex && i < len(m.Metadata.MetadataItems); i++ {
		if i != m.Metadata.RenderIndex {
			view += "\n"
		}
		data := truncateMiddleText(m.Metadata.MetadataItems[i][1], valueLength, "...")
		metadataName := m.Metadata.MetadataItems[i][0]
		if footerWidth(m.Context.WindowWidth)-maxKeyLength-3 < footerWidth(m.Context.WindowWidth)/2 {
			metadataName = truncateMiddleText(m.Metadata.MetadataItems[i][0], valueLength, "...")
		}
		view += fmt.Sprintf("%-*s %s", sprintfLength, metadataName, data)

	}

	borderWidth := footerWidth(m.Context.WindowWidth)
	borderHight := bottomElementHeight(bottomHeight)

	border := NewCard(borderWidth, borderHight, "Metadata", view, fmt.Sprintf("%s/%s", strconv.Itoa(m.Metadata.RenderIndex+1), strconv.Itoa(len(m.Metadata.MetadataItems))), "m", m.Context.FocusPanel == MetadataFocus)
	view = border.View()
	return view
}

func (m Model) clipboardRender() string {
	view := ""
	if len(m.copyItems.items) == 0 {
		view += "\n " + icon.Error + "  No content in clipboard"
	} else {
		for i := 0; i < len(m.copyItems.items) && i < bottomElementHeight(bottomHeight); i++ {
			if i == bottomElementHeight(bottomHeight)-1 {
				view += strconv.Itoa(len(m.copyItems.items)-i+1) + " item left...."
			} else {
				fileInfo, err := os.Stat(m.copyItems.items[i])
				if err != nil {
					outPutLog("Clipboard render function get item state error", err)
				}
				if !os.IsNotExist(err) {
					view += clipboardPrettierName(m.copyItems.items[i], footerWidth(m.Context.WindowWidth)-3, fileInfo.IsDir(), false) + "\n"
				}
			}
		}
	}

	borderWidth := 0
	if m.Context.WindowWidth%3 != 0 {
		borderWidth = footerWidth(m.Context.WindowWidth + m.Context.WindowWidth%3 + 2)
	} else {
		borderWidth = footerWidth(m.Context.WindowWidth)
	}
	borderHight := bottomElementHeight(bottomHeight)

	border := NewCard(borderWidth, borderHight, "Clipboard", view, "", "", false)
	view = border.View()
	return view
}

func (m Model) terminalSizeWarnRender() string {
	fullWidthString := strconv.Itoa(m.Context.WindowWidth)
	fullHeightString := strconv.Itoa(m.Context.WindowHeight)
	minimumWidthString := strconv.Itoa(minimumWidth)
	minimumHeightString := strconv.Itoa(minimumHeight)
	if m.Context.WindowHeight < minimumHeight {
		fullHeightString = terminalTooSmall.Render(fullHeightString)
	}
	if m.Context.WindowWidth < minimumWidth {
		fullWidthString = terminalTooSmall.Render(fullWidthString)
	}
	fullHeightString = terminalCorrectSize.Render(fullHeightString)
	fullWidthString = terminalCorrectSize.Render(fullWidthString)

	heightString := mainStyle.Render(" Height = ")
	return GetFullScreenStyle(m.Context.WindowHeight, m.Context.WindowWidth).Render(`Terminal size too small:` + "\n" +
		"Width = " + fullWidthString +
		heightString + fullHeightString + "\n\n" +

		"Needed for current config:" + "\n" +
		"Width = " + terminalCorrectSize.Render(minimumWidthString) +
		heightString + terminalCorrectSize.Render(minimumHeightString))
}

func (m Model) terminalSizeWarnAfterFirstRender() string {
	minimumWidthInt := Config.SidebarWidth + 20*len(m.File.FilePanels) + 20 - 1
	minimumWidthString := strconv.Itoa(minimumWidthInt)
	fullWidthString := strconv.Itoa(m.Context.WindowWidth)
	fullHeightString := strconv.Itoa(m.Context.WindowHeight)
	minimumHeightString := strconv.Itoa(minimumHeight)

	if m.Context.WindowHeight < minimumHeight {
		fullHeightString = terminalTooSmall.Render(fullHeightString)
	}
	if m.Context.WindowWidth < minimumWidthInt {
		fullWidthString = terminalTooSmall.Render(fullWidthString)
	}
	fullHeightString = terminalCorrectSize.Render(fullHeightString)
	fullWidthString = terminalCorrectSize.Render(fullWidthString)

	heightString := mainStyle.Render(" Height = ")
	return GetFullScreenStyle(m.Context.WindowHeight, m.Context.WindowWidth).Render(`You change your terminal size too small:` + "\n" +
		"Width = " + fullWidthString +
		heightString + fullHeightString + "\n\n" +

		"Needed for current config:" + "\n" +
		"Width = " + terminalCorrectSize.Render(minimumWidthString) +
		heightString + terminalCorrectSize.Render(minimumHeightString))
}

func (m Model) warnModalRender() string {
	title := m.warnModal.title
	content := m.warnModal.content
	confirm := modalConfirm.Render(" (" + hotkeys.Confirm[0] + ") Confirm ")
	cancel := modalCancel.Render(" (" + hotkeys.Quit[0] + ") Cancel ")
	tip := confirm + lipgloss.NewStyle().Background(modalBGColor).Render("           ") + cancel
	return modalBorderStyle(modalHeight, modalWidth).Render(title + "\n\n" + content + "\n\n" + tip)
}

func (m Model) commandLineInputBoxRender() string {
	return m.commandLine.input.View()
}
