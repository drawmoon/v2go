package internal

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/term/ansi"
)

func truncateText(text string, maxChars int, talis string) string {

	truncatedText := ansi.Truncate(text, maxChars-len(talis), "")
	if text != truncatedText {
		return truncatedText + talis
	}

	return text
}

func truncateTextBeginning(text string, maxChars int, talis string) string {
	if ansi.StringWidth(text) <= maxChars {
		return text
	}

	truncatedRunes := []rune(text)

	truncatedWidth := ansi.StringWidth(string(truncatedRunes))

	for truncatedWidth > maxChars {
		truncatedRunes = truncatedRunes[1:]
		truncatedWidth = ansi.StringWidth(string(truncatedRunes))
	}

	if len(truncatedRunes) > len(talis) {
		truncatedRunes = append([]rune(talis), truncatedRunes[len(talis):]...)
	}

	return string(truncatedRunes)
}

func truncateMiddleText(text string, maxChars int, talis string) string {
	if utf8.RuneCountInString(text) <= maxChars {
		return text
	}

	halfEllipsisLength := (maxChars - 3) / 2

	truncatedText := text[:halfEllipsisLength] + talis + text[utf8.RuneCountInString(text)-halfEllipsisLength:]

	return truncatedText
}

func prettierName(name string, width int, isDir bool, isSelected bool, bgColor lipgloss.Color) string {
	if isSelected {
		return filePanelItemSelectedStyle.Render(truncateText(name, width, "..."))
	} else {
		return filePanelStyle.Render(truncateText(name, width, "..."))
	}
}

func clipboardPrettierName(name string, width int, isDir bool, isSelected bool) string {
	style := getElementIcon(name, isDir)
	if isSelected {
		return GetColorStyle(lipgloss.Color(style.Color), footerBGColor).
			Background(footerBGColor).
			Render(style.Icon+" ") +
			filePanelItemSelectedStyle.Render(truncateTextBeginning(name, width, "..."))
	} else {
		return GetColorStyle(lipgloss.Color(style.Color), footerBGColor).
			Background(footerBGColor).
			Render(style.Icon+" ") +
			filePanelStyle.Render(truncateTextBeginning(name, width, "..."))
	}
}

func fileNameWithoutExtension(fileName string) string {
	for {
		pos := strings.LastIndexByte(fileName, '.')
		if pos <= 0 {
			break
		}
		fileName = fileName[:pos]
	}
	return fileName
}

func formatFileSize(size int64) string {
	if size == 0 {
		return "0B"
	}

	unitsDec := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	unitsBin := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

	if Config.FileSizeUseSI {
		unitIndex := int(math.Floor(math.Log(float64(size)) / math.Log(1000)))
		adjustedSize := float64(size) / math.Pow(1000, float64(unitIndex))
		return fmt.Sprintf("%.2f %s", adjustedSize, unitsDec[unitIndex])
	} else {
		unitIndex := int(math.Floor(math.Log(float64(size)) / math.Log(1024)))
		adjustedSize := float64(size) / math.Pow(1024, float64(unitIndex))
		return fmt.Sprintf("%.2f %s", adjustedSize, unitsBin[unitIndex])
	}
}
