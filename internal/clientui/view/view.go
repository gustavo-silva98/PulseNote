package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gustavo-silva98/adnotes/internal/clientui/model"
)

//var termWid, termHeight, _ = term.GetSize(os.Stdout.Fd())

var logoLines = []string{
	" _____     _         _____     _       ",
	"|  _  |_ _| |___ ___|   | |___| |_ ___ ",
	"|   __| | | |_ -| -_| | | | _ |  _| -_|",
	"|  |  | | | |   |   | |   |   | | |   |",
	"|__|  |___|_|___|___|_|___|___|_| |___|",
}

var gradientColors = []string{
	"#6d40f3ff", // roxo profundo
	"#7e40faff", // lilás mais vivo
	"#8b3bfcff", // lavanda
	"#BC78FE",   // rosa arroxeado (nova cor)
	"#B262FD",   // violeta claro
}

func renderLogo() string {
	var rendered string
	for i, line := range logoLines {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(gradientColors[i])).
			Align(lipgloss.Center)
		rendered += style.Render(line) + "\n"
	}
	return fmt.Sprintf("\n\n%v", rendered)
}

func View(m model.Model) string {
	var output string
	if m.Quitting {
		return "Bye!\n"
	}

	switch m.State {
	case model.InsertNoteState:
		output = InsertNoteView(m)
	case model.ReadNotesState:
		output = EditNoteView(m)
	case model.EditNoteSate:
		output = EditNoteView(m)
	case model.ConfirmEditSate:
		output = EditNoteView(m) + YesNoModalOverlay(m, "Do you want to save changes?")
	case model.DeleteNoteState:
		output = EditNoteView(m) + YesNoModalOverlay(m, "Do you want to delete?")
	case model.ResultEditState:
		output = EditNoteView(m) + ResultEditModalOverlay(m, m.ResultMessage)
	case model.ConfirmKillServerState:
		output = YesNoModalOverlay(m, m.ResultMessage)
	case model.FinishServerState:
		output = ResultEditModalOverlay(m, m.ResultMessage)
	case model.InitServerState:
		output = InitServerView(m)
	case model.FullSearchNoteState:
		output = FullSearchNoteView(m)
	case model.SaveNewNoteState:
		output = ResultEditModalOverlay(m, m.ResultMessage)
	}

	return output
}

func InsertNoteView(m model.Model) string {

	logoHeight := m.TermHeight / 3
	textHeight := m.TermHeight / 2
	helpheight := m.TermHeight - logoHeight - textHeight
	elementWidth := m.TermWidth - (m.TermWidth / 10)

	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6d40f3ff")).
		Align(lipgloss.Center).
		Width(elementWidth).
		Height(logoHeight)

	var textStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7e40fa")).
		Width(elementWidth).
		Height(textHeight)

	var helpStyle = lipgloss.NewStyle().
		AlignVertical(lipgloss.Bottom).
		AlignHorizontal(lipgloss.Center).
		Height(helpheight).
		Width(elementWidth)

	content := fmt.Sprintf(
		"Digite sua anotação abaixo. \n\n%s",
		m.Textarea.View(),
	)

	mainContent := lipgloss.JoinVertical(
		lipgloss.Top,
		logoStyle.Render(renderLogo()),
		textStyle.Render(content),
		helpStyle.Render(m.Help.ShortHelpView(m.HelpKeys)),
	)

	output := lipgloss.Place(
		m.TermWidth,
		m.TermHeight,
		lipgloss.Center, lipgloss.Top,
		mainContent,
	)

	return output
}

func InitServerView(m model.Model) string {
	logoHeight := (m.TermHeight / 10) * 6
	textHeight := m.TermHeight - logoHeight
	helpheight := m.TermHeight - logoHeight - textHeight
	elementWidth := m.TermWidth - (m.TermWidth / 10)

	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6d40f3ff")).
		Align(lipgloss.Center).
		Width(elementWidth).
		Height(logoHeight)

	var textStyle = lipgloss.NewStyle().
		Width(elementWidth).
		Height(textHeight).
		AlignHorizontal(lipgloss.Center).
		Foreground(lipgloss.Color("#909090"))

	options := []string{
		"Ctrl + Shift + H -> Save Note",
		"Ctrl + Shift + R -> Read Note",
		"Ctrl + Shift + K -> Kill Server",
		"Ctrl + Shift + D -> Advanced Search",
	}
	options = KeysForInitState(options, 20)

	optionsFormatted := SliceFormatter(options)
	optionsFormatted = append([]string{"HotKeys"}, optionsFormatted...)

	content := strings.Join(optionsFormatted, "\n\n\n")

	var helpStyle = lipgloss.NewStyle().
		AlignVertical(lipgloss.Bottom).
		AlignHorizontal(lipgloss.Center).
		Height(helpheight).
		Width(elementWidth)

	mainContent := lipgloss.JoinVertical(
		lipgloss.Top,
		logoStyle.Render(renderLogo()),
		textStyle.Render(content),
		helpStyle.Render(m.Help.ShortHelpView(m.HelpKeys)),
	)

	output := lipgloss.Place(
		m.TermWidth,
		m.TermHeight,
		lipgloss.Center, lipgloss.Top,
		mainContent,
	)

	return output
}

func textareaEditView(m model.Model) string {
	var textStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7e40faff"))

	return textStyle.Render(m.TextareaEdit.View())
}

func EditNoteView(m model.Model) string {
	listWidth := m.TermWidth / 2
	editorWidth := m.TermWidth - listWidth

	contentHeight := m.TermHeight - 3
	helpHeight := 3

	listStyle := lipgloss.NewStyle().
		Width(listWidth).
		Height(contentHeight)

	editorStyle := lipgloss.NewStyle().
		Width(editorWidth).
		Height(contentHeight)

	helpStyle := lipgloss.NewStyle().
		AlignVertical(lipgloss.Bottom).
		AlignHorizontal(lipgloss.Center).
		Width(m.TermWidth).
		Height(helpHeight)

	list := listStyle.Render(m.ListModel.View())
	editor := editorStyle.Render(textareaEditView(m))
	horizontal := lipgloss.JoinHorizontal(lipgloss.Top, list, editor)
	output := lipgloss.JoinVertical(lipgloss.Top, horizontal, helpStyle.Render(m.Help.ShortHelpView(m.HelpKeys)))

	return output
}

func ListModelView(m model.Model) string {
	var listModelStyle = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center)

	return listModelStyle.Render(m.ListModel.View())
}

func YesNoModalOverlay(m model.Model, question string) string {
	overlay := lipgloss.NewStyle().
		Width(m.TermWidth).
		Height(m.TermHeight).
		Faint(true).
		Render(strings.Repeat(" ", m.TermWidth*m.TermHeight/2))

	modalWidth := m.TermWidth / 3
	if modalWidth > m.TermWidth {
		modalWidth = m.TermWidth
	}

	modalHeight := 7
	if modalHeight > m.TermHeight {
		modalHeight = m.TermHeight
	}

	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7e40fa")).
		Background(lipgloss.Color("#22223b"))

	questionStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#fff")).
		PaddingTop(2)

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#fff")).
		Background(lipgloss.Color("#7e40fa")).
		Padding(0, 2).
		Margin(1, 1)

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		buttonStyle.Render("[Y]es"),
		buttonStyle.Render("[N]o"),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		questionStyle.Render(question),
		buttons,
	)

	modal := lipgloss.Place(
		m.TermWidth, m.TermHeight,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(content),
	)

	return overlay + modal
}

func ResultEditModalOverlay(m model.Model, question string) string {
	overlay := lipgloss.NewStyle().
		Width(m.TermWidth).
		Height(m.TermHeight).
		Faint(true).
		Render(strings.Repeat(" ", m.TermWidth*m.TermHeight/2))

	modalWidth := m.TermWidth / 3
	if modalWidth > m.TermWidth {
		modalWidth = m.TermWidth
	}

	modalHeight := 7
	if modalHeight > m.TermHeight {
		modalHeight = m.TermHeight
	}

	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7e40fa")).
		Background(lipgloss.Color("#22223b"))

	questionStyle := lipgloss.NewStyle().
		Align(lipgloss.Bottom).
		PaddingTop(3).
		Foreground(lipgloss.Color("#fff"))

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		questionStyle.Render(question),
	)

	modal := lipgloss.Place(
		m.TermWidth, m.TermHeight,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(content),
	)

	return overlay + modal
}

func KeysForInitState(sliceKeys []string, totalLenght int) []string {
	for i, key := range sliceKeys {
		if i/2 > 0 {
			sliceKeys[i] = FormatRightString(key, totalLenght)
		}
		sliceKeys[i] = FormatLeftString(key, totalLenght)
	}
	return sliceKeys
}

func FormatRightString(text string, lenght int) string {
	if len(text) >= lenght {
		return text
	}
	space := strings.Repeat(" ", lenght-len(text))
	return fmt.Sprintf("%v%v", space, text)
}

func FormatLeftString(text string, lenght int) string {
	if len(text) >= lenght {
		return text
	}
	space := strings.Repeat(" ", lenght-len(text))
	return fmt.Sprintf("%v%v", text, space)
}

func FormatCenterString(text string, lenght int) string {
	if len(text) >= lenght {
		return text
	}

	space := lenght - len(text)
	leftSide := space / 2
	rightSide := space - leftSide

	return fmt.Sprintf("%v%v%v", strings.Repeat(" ", leftSide), text, strings.Repeat(" ", rightSide))
}

func SliceFormatter(sliceIn []string) []string {
	var sliceResult []string
	for i := 0; i <= len(sliceIn)-1; i += 2 {
		if i+1 > len(sliceIn)-1 {
			sliceResult = append(sliceResult, sliceIn[i])
		} else {
			sliceResult = append(sliceResult, fmt.Sprintf("%v   %v", sliceIn[i], sliceIn[i+1]))
		}
	}
	return sliceResult
}

func FullSearchNoteView(m model.Model) string {
	listWidth := m.TermWidth / 2
	editorWidth := listWidth - listWidth/10

	searchBoxHeight := 1
	listHeight := 5

	editBoxHeight := m.TermHeight - 3
	helpHeight := 3

	var textStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7e40fa")).
		Width(listWidth).
		Height(searchBoxHeight).
		MarginBottom(1)

	listStyle := lipgloss.NewStyle().
		Width(listWidth).
		Height(listHeight)

	editorStyle := lipgloss.NewStyle().
		Width(editorWidth).
		Height(editBoxHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7e40faff"))

	helpStyle := lipgloss.NewStyle().
		AlignVertical(lipgloss.Bottom).
		AlignHorizontal(lipgloss.Center).
		Width(m.TermWidth).
		Height(helpHeight)

	searchBox := textStyle.Render(m.TextAreaSearch.View())
	list := listStyle.Render(m.ListModel.View())
	editor := editorStyle.Render(m.TextareaEdit.View())
	leftSide := lipgloss.JoinVertical(lipgloss.Top, searchBox, list)
	horizontal := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, editor)
	output := lipgloss.JoinVertical(lipgloss.Top, horizontal, helpStyle.Render(m.Help.ShortHelpView(m.HelpKeys)))

	return output
}
