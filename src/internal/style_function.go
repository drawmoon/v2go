package internal

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/yorukot/superfile/src/config/icon"
)

// Generate modal (pop up widnwos) border style
func modalBorderStyle(height int, width int) lipgloss.Style {
	border := GetBorder()
	return lipgloss.NewStyle().Height(height).
		Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Border(border).
		BorderForeground(footerBorderActiveColor).
		BorderBackground(modalBGColor).
		Background(modalBGColor).
		Foreground(modalFGColor)
}

// Generate full screen style for terminal size too small etc
func GetFullScreenStyle(height int, width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Height(height).
		Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Background(fullScreenBGColor).
		Foreground(fullScreenFGColor)
}

// Return only fg and bg color style
func GetColorStyle(fgColor lipgloss.Color, bgColor lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(fgColor).
		Background(bgColor)
}

// Generate config error style
func loadConfigError(value string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("■ ERROR: ") + "Config file \"" + lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Render(value) + "\" invalidation"
}

// Generate config error style
func lodaHotkeysError(value string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("■ ERROR: ") + "Hotkeys file \"" + lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Render(value) + "\" invalidation"
}

// Generate command line in the bottom
func generateCommandLineInputBox() textinput.Model {
	ti := textinput.New()
	ti.Cursor.Style = footerCursorStyle
	ti.Cursor.TextStyle = footerStyle
	ti.TextStyle = filePanelStyle
	ti.Prompt = filePanelTopDirectoryIconStyle.Render(icon.Cursor + icon.Space)
	ti.Cursor.Blink = true
	ti.PlaceholderStyle = filePanelStyle
	ti.Blur()
	return ti
}
