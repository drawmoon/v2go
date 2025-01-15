package internal

import (
	"os"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/yorukot/superfile/src/config/icon"
)

// Model for sidebar
type SidebarModel struct {
	Title       string
	RenderIndex int
	Cursor      int
	Directories []Directory
}

type Directory struct {
	Name     string
	Location string
}

func NewSidebarModel() SidebarModel {
	return SidebarModel{
		Title:       "Superfile",
		RenderIndex: 0,
		Directories: getDirectories(),
	}
}

func (m Model) renderSidebar() string {
	if Config.SidebarWidth == 0 {
		return ""
	}
	view := ansi.Truncate(sidebarTitleStyle.Render(m.Sidebar.Title), Config.SidebarWidth, "")
	view += "\n"

	totalHeight := 2
	for i := m.Sidebar.RenderIndex; i < len(m.Sidebar.Directories); i++ {
		if totalHeight >= m.mainPanelHeight {
			break
		} else {
			view += "\n"
		}

		directory := m.Sidebar.Directories[i]

		totalHeight++
		cursor := " "
		if m.Sidebar.Cursor == i && m.Context.FocusPanel == SidebarFocus {
			cursor = icon.Cursor
		}

		if directory.Location == m.File.FilePanels[m.filePanelFocusIndex].Location {
			view += filePanelCursorStyle.Render(cursor+" ") + sidebarSelectedStyle.Render(truncateText(directory.Name, Config.SidebarWidth-2, "..."))
		} else {
			view += filePanelCursorStyle.Render(cursor+" ") + sidebarStyle.Render(truncateText(directory.Name, Config.SidebarWidth-2, "..."))
		}
	}

	border := NewCard(Config.SidebarWidth, m.mainPanelHeight, "Superfile", view, "", "s", m.Context.FocusPanel == SidebarFocus)
	view = border.View()
	return view
}

// Return all sidebar directories
func getDirectories() []Directory {
	directories := []Directory{}

	// Return system default directory e.g. Home, Downloads, etc
	tmp := []Directory{
		{Location: xdg.Home, Name: "Home"},
		{Location: xdg.UserDirs.Download, Name: "Downloads"},
		{Location: xdg.UserDirs.Documents, Name: "Documents"},
		{Location: xdg.UserDirs.Pictures, Name: "Pictures"},
		{Location: xdg.UserDirs.Videos, Name: "Videos"},
		{Location: xdg.UserDirs.Music, Name: "Music"},
		{Location: xdg.UserDirs.Templates, Name: "Templates"},
		{Location: xdg.UserDirs.PublicShare, Name: "PublicShare"},
	}
	for _, dir := range tmp {
		if _, err := os.Stat(dir.Location); !os.IsNotExist(err) {
			// Directory exists
			directories = append(directories, dir)
		}
	}

	return directories
}
