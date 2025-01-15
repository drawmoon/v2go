package app

import (
	"embed"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	internal "github.com/yorukot/superfile/src/internal"
)

// Run application
func Run(content embed.FS) {
	internal.InitConfigFile(content)

	p := tea.NewProgram(internal.NewModel("", false), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
