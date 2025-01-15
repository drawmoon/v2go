package internal

import (
	"fmt"

	"github.com/yorukot/superfile/src/config/icon"
	stringfunction "github.com/yorukot/superfile/src/pkg/string_function"
)

type MenuModal struct {
	Width       int
	Height      int
	Focus       bool
	RenderIndex int
	Cursor      int
	MenuItems   []MenuItem
}

type MenuItem struct {
	SubTitle       string
	Description    string
	Hotkey         []string
	HotkeyWorkType hotkeyType
}

func NewMenuModel() MenuModal {
	menuData := []MenuItem{}
	return MenuModal{
		RenderIndex: 0,
		Cursor:      1,
		MenuItems:   menuData,
		Focus:       false,
	}
}

func (m Model) prepareView(bg string) string {
	view := ""
	maxKeyLength := 0

	for _, item := range m.Menu.MenuItems {
		totalKeyLen := 0
		for _, key := range item.Hotkey {
			totalKeyLen += len(key)
		}
		saprateLen := len(item.Hotkey) - 1*3
		if item.SubTitle == "" && totalKeyLen+saprateLen > maxKeyLength {
			maxKeyLength = totalKeyLen + saprateLen
		}
	}

	valueLength := m.Menu.Width - maxKeyLength - 2
	if valueLength < m.Menu.Width/2 {
		valueLength = m.Menu.Width/2 - 2
	}

	renderHotkeyLength := 0
	totalTitleCount := 0
	cursorBeenTitleCount := 0

	for i, data := range m.Menu.MenuItems {
		if data.SubTitle != "" {
			if i < m.Menu.Cursor {
				cursorBeenTitleCount++
			}
			totalTitleCount++
		}
	}

	for i := m.Menu.RenderIndex; i < m.Menu.Height+m.Menu.RenderIndex && i < len(m.Menu.MenuItems); i++ {
		hotkey := ""
		if m.Menu.MenuItems[i].SubTitle != "" {
			continue
		}

		for i, key := range m.Menu.MenuItems[i].Hotkey {
			if i != 0 {
				hotkey += " | "
			}
			hotkey += key
		}

		if len(helpMenuHotkeyStyle.Render(hotkey)) > renderHotkeyLength {
			renderHotkeyLength = len(helpMenuHotkeyStyle.Render(hotkey))
		}
	}

	for i := m.Menu.RenderIndex; i < m.Menu.Height+m.Menu.RenderIndex && i < len(m.Menu.MenuItems); i++ {
		if i != m.Menu.RenderIndex {
			view += "\n"
		}

		if m.Menu.MenuItems[i].SubTitle != "" {
			view += helpMenuTitleStyle.Render(" " + m.Menu.MenuItems[i].SubTitle)
			continue
		}

		hotkey := ""
		description := truncateText(m.Menu.MenuItems[i].Description, valueLength, "...")

		for i, key := range m.Menu.MenuItems[i].Hotkey {
			if i != 0 {
				hotkey += " | "
			}
			hotkey += key
		}

		cursor := "  "
		if m.Menu.Cursor == i {
			cursor = filePanelCursorStyle.Render(icon.Cursor + " ")
		}
		view += cursor + modalStyle.Render(fmt.Sprintf("%*s%s", renderHotkeyLength, helpMenuHotkeyStyle.Render(hotkey+" "), modalStyle.Render(description)))
	}

	border := NewCard(m.Menu.Width, m.Menu.Height, "Help", view, "", "", true)
	view = border.View()

	overlayX := m.Context.WindowWidth/2 - m.Menu.Width/2
	overlayY := m.Context.WindowHeight/2 - m.Menu.Height/2
	return stringfunction.PlaceOverlay(overlayX, overlayY, view, bg)
}
