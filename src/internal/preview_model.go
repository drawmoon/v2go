package internal

type PreviewModel struct {
	Width int
	Open  bool
}

func NewPreviewModel() PreviewModel {
	return PreviewModel{
		Open: Config.DefaultOpenFilePreview,
	}
}

func (m Model) renderPreview() string {
	m.File.Preview.Width = m.Context.WindowWidth - Config.SidebarWidth - m.File.Width - 6

	border := NewCard(m.File.Preview.Width, m.mainPanelHeight, "Preview", "", "", "", false)
	return border.View()
}
