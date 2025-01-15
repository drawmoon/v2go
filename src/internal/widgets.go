package internal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// >>>>>>>>>>>>>>>> Card >>>>>>>>>>>>>>>>

type Card struct {
	Width   int
	Height  int
	Title   string
	Content string
	Footer  string
	Hotkey  string
	Active  bool
}

func NewCard(width, height int, title, content, footer, hotkey string, active bool) Card {
	return Card{
		Width:   width,
		Height:  height,
		Title:   title,
		Content: content,
		Footer:  footer,
		Hotkey:  hotkey,
		Active:  active,
	}
}

func (b *Card) View() string {
	border := GetBorder()
	if b.Title != "" {
		border.Top = GetTopBorderView(b.Title, b.Hotkey, b.Width)
	}
	if b.Footer != "" {
		border.Bottom = GetBottomBorderView(b.Footer, b.Width)
	}

	borderColor := footerBorderColor
	if b.Active {
		borderColor = footerBorderActiveColor
	}

	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(borderColor).
		BorderBackground(footerBGColor).
		Width(b.Width).
		Height(b.Height).
		Background(footerBGColor).
		Foreground(footerFGColor).
		Render(b.Content)
}

// Get top border display strings.
func GetTopBorderView(title, hotkey string, width int) string {
	if hotkey != "" {
		title = fmt.Sprintf("[%s] %s", hotkey, title)
	}
	strs := Config.BorderTop + " " + title + " "
	return strs + strings.Repeat(Config.BorderTop, width-len(strs)+1)
}

// Get bottom border display strings.
func GetBottomBorderView(title string, width int) string {
	size := len(title)
	if size == 0 {
		return strings.Repeat(Config.BorderBottom, width)
	}
	strs := Config.BorderBottom + " " + title + " "
	return strings.Repeat(Config.BorderBottom, width-len(strs)+1) + strs
}

func GetBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         Config.BorderTop,
		Bottom:      Config.BorderBottom,
		Left:        Config.BorderLeft,
		Right:       Config.BorderRight,
		TopLeft:     Config.BorderTopLeft,
		TopRight:    Config.BorderTopRight,
		BottomLeft:  Config.BorderBottomLeft,
		BottomRight: Config.BorderBottomRight,
	}
}

// <<<<<<<<<<<<<<<< Card <<<<<<<<<<<<<<<<

// >>>>>>>>>>>>>>>> SearchBar >>>>>>>>>>>>>>>>

type SearchBar struct {
}

func GetSearchModel() textinput.Model {
	input := textinput.New()

	input.Cursor.Style = lipgloss.NewStyle().Foreground(cursorColor).Background(footerBGColor)
	input.Cursor.TextStyle = lipgloss.NewStyle().Foreground(footerFGColor).Background(footerBGColor)
	input.Cursor.Blink = true

	input.TextStyle = filePanelStyle
	input.PlaceholderStyle = filePanelStyle
	input.Placeholder = "[" + hotkeys.SearchBar[0] + "] Type something"

	input.Blur()
	return input
}

// <<<<<<<<<<<<<<<< SearchBar <<<<<<<<<<<<<<<<

// >>>>>>>>>>>>>>>> Cursor >>>>>>>>>>>>>>>>

type Cursor struct {
	Icon      string
	Style     lipgloss.Style
	TextStyle lipgloss.Style
}

func (c *Cursor) View() string {
	return c.Style.Render(c.Icon + " ")
}

func NewCursor() Cursor {
	return Cursor{
		Icon:      "→",
		Style:     lipgloss.NewStyle().Foreground(cursorColor).Background(filePanelBGColor),
		TextStyle: lipgloss.NewStyle().Foreground(footerFGColor).Background(footerBGColor),
	}
}

// <<<<<<<<<<<<<<<< Cursor <<<<<<<<<<<<<<<<
