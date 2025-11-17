package update

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gustavo-silva98/adnotes/internal/clientui/model"
	"github.com/gustavo-silva98/adnotes/internal/repository/file"
	"github.com/muesli/reflow/wordwrap"
)

// var termWidth, termHeight, _ = term.GetSize(os.Stdout.Fd())
var ctx = context.Background()

const PageSize = 10

// Mensagem para timeout do resultado da edição
type resultEditTimeoutMsg struct{}
type resultKillTimeoutMsg struct{}
type fullSearchDebounceMsg struct{}
type resultSaveNewNote struct{}
type noteItem struct {
	title, desc  string
	NoteText     string
	Id           int
	Reminder     int
	PlusReminder int
}

func (i noteItem) Title() string       { return i.title }
func (i noteItem) Description() string { return i.desc }
func (i noteItem) FilterValue() string { return i.title }
func (i noteItem) IdValue() int        { return i.Id }

func Update(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TermHeight = msg.Height - msg.Height/10
		m.TermWidth = msg.Width

	case resultEditTimeoutMsg:
		m.State = model.ReadNotesState
		m.HelpKeys = helpMaker(m)

		return *m, nil
	case resultKillTimeoutMsg:
		return *m, tea.Quit
	case resultSaveNewNote:
		return *m, tea.Quit
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Read):
			m.State = model.ReadNotesState
		}
	}
	m.HelpKeys = helpMaker(m)

	switch m.State {
	case model.InsertNoteState:
		return updateInsertNoteState(msg, m)
	case model.ReadNotesState:
		return updateReadNoteState(msg, m)
	case model.EditNoteSate:
		return updateEditNoteFunc(msg, m)
	case model.ConfirmEditSate:
		return updateConfirmEditNote(msg, m)
	case model.DeleteNoteState:
		return updateConfirmDeleteNote(msg, m)
	case model.ResultEditState:
		m.State = model.ReadNotesState
	case model.ConfirmKillServerState:
		return UpdateConfirmKillServerState(msg, m)
	case model.InitServerState:
		return UpdateInitServerState(msg, m)
	case model.FullSearchNoteState:
		return UpdateSearchNotes(msg, m)
	case model.SaveNewNoteState:
		return updateResultSaveNewNote(msg, m)
	}
	return *m, nil
}

func UpdateConfirmKillServerState(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	cmds = append(cmds, cmd)
	m.ResultMessage = "Do you wanna terminate the server?"
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Yes):
			err := KillProcess("server")
			if err != nil {
				panic(fmt.Sprintf("erro ao finalizar Server %v", err))
			}
			m.State = model.FinishServerState
			m.ResultMessage = "Server terminated"
			return updateResultKillServerState(msg, m)
		case key.Matches(msg, m.Keys.No):
			m.Quitting = true
			return *m, tea.Quit
		}
	}
	return *m, tea.Batch(cmds...)
}

func updateResultKillServerState(_ tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	// retorna o cmd que vai enviar resultEditTimeoutMsg após 500ms
	return *m, tea.Tick(1000*time.Millisecond, func(t time.Time) tea.Msg {
		return resultKillTimeoutMsg{}
	})
}

func updateResultSaveNewNote(_ tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	// retorna o cmd que vai enviar resultEditTimeoutMsg após 500ms
	return *m, tea.Tick(1000*time.Millisecond, func(t time.Time) tea.Msg {
		return resultSaveNewNote{}
	})
}

func UpdateInitServerState(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			return *m, tea.Quit
		}
	}
	return *m, tea.Batch(cmds...)
}

func updateInsertNoteState(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.Textarea.SetWidth(m.TermWidth - (m.TermWidth / 10))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Save):
			noteExample := file.Note{
				Hour:         (time.Now().Unix() - int64(time.Now().Second())),
				NoteText:     m.Textarea.Value(),
				Reminder:     0,
				PlusReminder: 0,
			}

			_, err := m.DB.InsertNote(&noteExample, ctx)
			if err != nil {
				file.WriteLog(err.Error(), m.LogPath)
			}
			m.ResultMessage = "Note saved successfully!"
			m.State = model.SaveNewNoteState
		case key.Matches(msg, m.Keys.Esc):
			if m.Textarea.Focused() {
				m.Textarea.Blur()
			}
		case key.Matches(msg, m.Keys.FullSearch):
			m.State = model.FullSearchNoteState
			m.TextAreaSearch.SetWidth(m.TermWidth/2 - 4)
			m.TextAreaSearch.SetHeight(1)
			m.TextareaEdit.SetHeight(m.TermHeight - 5)
			m.TextareaEdit.SetWidth(m.TermWidth - m.ListModel.Width() - 4)
		case key.Matches(msg, m.Keys.Quit):
			m.Quitting = true
			return *m, tea.Quit
		case key.Matches(msg, m.Keys.Read):
			m.ItemList = queryMapNotes(m)
			m.State = model.ReadNotesState
			return *m, nil
		default:
			if !m.Textarea.Focused() {
				cmd = m.Textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	}
	m.Textarea, cmd = m.Textarea.Update(msg)
	cmds = append(cmds, cmd)

	return *m, tea.Batch(cmds...)

}

func updateReadNoteState(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	totalPages, hasNextPage, hasPrevPage := getPaginationInfo(m)
	paginateUp := false
	paginateDown := false

	file.WriteLog(fmt.Sprintf("Current Page = %v", m.CurrentPage), "")
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.Keys.Down) {
			if m.ListModel.Index() == PageSize-1 {
				file.WriteLog("Apertou Down", "")
				file.WriteLog(fmt.Sprintf("Current Page = %v", m.CurrentPage), "")
				if hasNextPage {
					m.CurrentPage += 1

					// Atualiza a lista de itens.
					m.ItemList = queryMapNotes(m)
					m.ListModel.SetItems(m.ItemList)
					paginateDown = true
				}
			}
		}
		if key.Matches(keyMsg, m.Keys.Up) {
			if m.ListModel.Index() == 0 && m.CurrentPage != 1 {
				file.WriteLog("Apertou Down", "")
				file.WriteLog(fmt.Sprintf("Current Page = %v", m.CurrentPage), "")
				if hasPrevPage {
					m.CurrentPage -= 1
				}

				// Atualiza a lista de itens.
				m.ItemList = queryMapNotes(m)
				m.ListModel.SetItems(m.ItemList)
				paginateUp = true
			}
		}
	}

	// Só recarregue a lista se ItemList estiver vazia (primeira vez) ou se mudar de página
	if len(m.ItemList) == 0 {
		m.ItemList = queryMapNotes(m)
		d := list.NewDefaultDelegate()
		c := lipgloss.Color("#FE02FF")
		c1 := lipgloss.Color("#7e40fa")
		d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(c).BorderLeftForeground(c).Bold(true)
		d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(lipgloss.Color("#9a6bf8ff")).Faint(true)
		d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(c1).BorderLeftForeground(c)
		d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(lipgloss.Color("#f2c9faff")).Faint(true)

		l := list.New(m.ItemList, d, m.TermWidth/3, m.TermHeight-3)
		l.Styles.Title = l.Styles.Title.Background(lipgloss.Color("#9D2EB0")).Foreground(lipgloss.Color("#E0D9F6"))
		l.SetShowHelp(false)

		m.ListModel = l
	}
	m.ListModel.SetSize(m.TermWidth/2, m.TermHeight-5)
	m.TextareaEdit.SetHeight(m.TermHeight - 5)
	m.TextareaEdit.SetWidth(m.TermWidth - m.ListModel.Width() - 2)
	m.ListModel.Title = fmt.Sprintf("Notas (%v/%v)", m.CurrentPage, totalPages)

	m.ListModel, cmd = m.ListModel.Update(msg)
	cmds = append(cmds, cmd)

	selected := m.ListModel.SelectedItem()
	if selected != nil {
		if note, ok := selected.(noteItem); ok {
			wrapped := wordwrap.String(fmt.Sprintf("%v", note.NoteText), m.TextareaEdit.Width())
			// Só atualize o valor se for diferente do atual
			if m.TextareaEdit.Value() != wrapped {
				m.TextareaEdit.SetValue(wrapped)
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			m.Quitting = true
			return *m, tea.Quit
		case key.Matches(msg, m.Keys.PageBack):
			m.State = model.InsertNoteState
		case key.Matches(msg, m.Keys.Delete):
			m.State = model.DeleteNoteState
		case key.Matches(msg, m.Keys.FullSearch):
			m.State = model.FullSearchNoteState
			m.TextAreaSearch.SetWidth(m.TermWidth/2 - 4)
			m.TextAreaSearch.SetHeight(1)
			m.ListModel.SetSize(m.TermWidth/2, m.TermHeight-5)
			m.TextareaEdit.SetHeight(m.TermHeight - 5)
			m.TextareaEdit.SetWidth(m.TermWidth - m.ListModel.Width() - 4)
		case key.Matches(msg, m.Keys.Enter):
			// Ao entrar no modo de edição, inicialize e foque o TextareaEdit
			m.State = model.EditNoteSate
			if !m.TextareaEdit.Focused() {
				cmd = m.TextareaEdit.Focus()
				cmds = append(cmds, cmd)
			}
		}
	}
	if paginateDown {
		m.ListModel.Select(0)
	}
	if paginateUp {
		m.ListModel.Select(PageSize - 1)
	}
	return *m, tea.Batch(cmds...)
}

func updateEditNoteFunc(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.ListModel.SetSize(m.TermWidth/2, m.TermHeight-5)
	m.TextareaEdit.SetHeight(m.TermHeight - 5)
	m.TextareaEdit.SetWidth(m.TermWidth - m.ListModel.Width() - 2)

	m.TextareaEdit, cmd = m.TextareaEdit.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			m.State = model.ReadNotesState
			if m.TextareaEdit.Focused() {
				m.TextareaEdit.Blur()
			}

		case key.Matches(msg, m.Keys.Save):
			m.State = model.ConfirmEditSate
		}

	}
	return *m, tea.Batch(cmds...)
}

func updateConfirmDeleteNote(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Yes):
			if selected := m.ListModel.SelectedItem(); selected != nil {
				if note, ok := selected.(noteItem); ok {
					rowsUpdated, err := m.DB.DeleteNoteRepository(ctx, note.Id)
					if err != nil {
						m.ResultMessage = fmt.Sprintf("Erro: %v\nErro ao deletar a nota.", err.Error())
						m.State = model.ReadNotesState
					}
					if rowsUpdated == 1 {
						m.ResultMessage = fmt.Sprintf("Nota %v deletada com sucesso.", note.title)
						m.ItemList = queryMapNotes(m)
						m.ListModel.SetItems(m.ItemList)
						m.State = model.ResultEditState
						return updateResultEditState(msg, m)
					}
				}
			}
		case key.Matches(msg, m.Keys.No):
			m.State = model.ReadNotesState
		}

	}
	return *m, tea.Batch(cmds...)
}

func updateConfirmEditNote(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd

	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Yes):
			if selected := m.ListModel.SelectedItem(); selected != nil {
				if note, ok := selected.(noteItem); ok {
					noteInput := file.Note{
						ID:           note.Id,
						Hour:         time.Now().Unix(),
						NoteText:     m.TextareaEdit.Value(),
						Reminder:     note.Reminder,
						PlusReminder: note.PlusReminder,
					}
					rowsUpdated, err := m.DB.UpdateEditNoteRepository(ctx, noteInput)
					if err != nil {
						m.ResultMessage = fmt.Sprintf("Erro: %v\nErro ao salvar a nota. Necessário averiguar.", err.Error())
						m.State = model.ResultEditState
					}
					if rowsUpdated == 1 {
						m.ResultMessage = fmt.Sprintf("Nota %v editada com sucesso.", note.title)
						m.ItemList = queryMapNotes(m)
						m.ListModel.SetItems(m.ItemList)
						m.State = model.ResultEditState
						return updateResultEditState(msg, m)

					}
				}
			}
		case key.Matches(msg, m.Keys.No):
			m.State = model.EditNoteSate
		}
	}
	return *m, tea.Batch(cmds...)
}

func updateResultEditState(_ tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	// retorna o cmd que vai enviar resultEditTimeoutMsg após 500ms
	return *m, tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		return resultEditTimeoutMsg{}
	})
}

func helpMaker(m *model.Model) []key.Binding {
	// helper pra formatar tecla+descrição
	b := func(keys, helpText string) key.Binding {
		return key.NewBinding(
			key.WithKeys(keys),
			key.WithHelp(keys, helpText),
		)
	}

	switch m.State {
	case model.InsertNoteState:
		return []key.Binding{
			b("Ctrl + s", "Save and Quit"),
			b("Ctrl + r", "Read Notes"),
			b("Ctrl + q", "Quit"),
			b("Ctrl + a", "Advanced Search"),
		}
	case model.ReadNotesState:
		return []key.Binding{
			b("Alt + ←", "Insert Note"),
			b("Enter", "Edit Note"),
			b("Ctrl + a", "Advanced Search"),
			b("Ctrl + d", "Delete Note"),
			b("Ctrl + q", "Quit"),
		}
	case model.EditNoteSate:
		return []key.Binding{
			b("Ctrl + s", "Save Note"),
			b("Ctrl + q", "Quit Editing"),
		}
	case model.InitServerState:
		return []key.Binding{
			b("Ctrl + q", "Close Window"),
		}
	case model.FullSearchNoteState:
		return []key.Binding{
			b("Ctrl + q", "Close Window"),
			b("Ctrl + r", "Read Notes"),
		}
	}
	return []key.Binding{}
}

func titleFormatter(title string) string {
	maxLineLenght := 40
	splitStr := strings.Split(title, ",")[0]
	splitStr = strings.Split(splitStr, "\n")[0]
	if len(splitStr) > maxLineLenght {
		return splitStr[0:maxLineLenght] + "..."
	}
	return splitStr
}

func KillProcess(processName string) error {
	switch runtime.GOOS {
	case "windows":
		processName += ".exe"
		return exec.Command("taskkill", "/IM", processName, "/F").Run()
	case "linux":
		return exec.Command("pkill", "-f", processName).Run()
	default:
		return fmt.Errorf("sistema operacional não suportado: %v", runtime.GOOS)
	}
}

func UpdateSearchNotes(msg tea.Msg, m *model.Model) (model.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	isNavigating := false
	oldValue := m.TextAreaSearch.Value()

	m.TextAreaSearch.SetWidth(m.TermWidth/2 - 4)
	m.TextAreaSearch.SetHeight(1)

	if len(m.ItemList) == 0 {
		d := list.NewDefaultDelegate()
		c := lipgloss.Color("#FE02FF")
		c1 := lipgloss.Color("#7e40fa")
		d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(c).BorderLeftForeground(c).Bold(true)
		d.Styles.NormalTitle = d.Styles.NormalTitle.Foreground(lipgloss.Color("#9a6bf8ff")).Faint(true)
		d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(c1).BorderLeftForeground(c)
		d.Styles.NormalDesc = d.Styles.NormalDesc.Foreground(lipgloss.Color("#f2c9faff")).Faint(true)
		l := list.New(m.ItemList, d, m.TermWidth/2, m.TermHeight-5)
		l.Styles.Title = l.Styles.Title.Background(lipgloss.Color("#9D2EB0")).Foreground(lipgloss.Color("#E0D9F6"))
		l.Title = "Resultados da Busca"
		l.SetShowHelp(false)
		m.ListModel = l
		m.ListModel.SetSize(m.TermWidth/2, m.TermHeight-5)
		m.TextareaEdit.SetHeight(m.TermHeight - 5)
		m.TextareaEdit.SetWidth(m.TermWidth - m.ListModel.Width() - 4)
	} else {

		m.ListModel.SetSize(m.TermWidth/2, m.TermHeight-5)
		m.TextareaEdit.SetHeight(m.TermHeight - 5)
		m.TextareaEdit.SetWidth(m.TermWidth - m.ListModel.Width() - 4)
	}

	if !m.FullSearchBool {
		m.TextAreaSearch.Placeholder = "Digite sua busca aqui..."
		m.TextAreaSearch.Focus()
		m.TextareaEdit.Blur()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			m.Quitting = true
			return *m, tea.Quit
		case key.Matches(msg, m.Keys.Read):
			m.State = model.ReadNotesState
		case key.Matches(msg, m.Keys.Up, m.Keys.Down):
			if !m.TextAreaSearch.Focused() {
				isNavigating = true
			}
		case key.Matches(msg, m.Keys.Enter):
			m.State = model.EditNoteSate
			if !m.TextareaEdit.Focused() {
				cmd = m.TextareaEdit.Focus()
				cmds = append(cmds, cmd)
			}
		default:
			if !isNavigating {
				m.FullSearchBool = false
				if m.TextAreaSearch.Value() != m.FullSearchQuery {
					m.FullSearchQuery = m.TextAreaSearch.Value()
					cmds = append(cmds, debouncerFullSearchNote(m, 500*time.Millisecond))
				}
				if !m.TextAreaSearch.Focused() {
					m.TextAreaSearch.Focus()
					m.TextareaEdit.SetValue("")
					m.TextareaEdit.Blur()
				}
			}
		}
	case fullSearchDebounceMsg:
		m.FullSearchBool = true
		m.FullSearchTimerCancel = nil
		m.ItemList = FullSearchQueryMapNotes(m)
		m.ListModel.SetItems(m.ItemList)
	}
	if !isNavigating {
		m.TextAreaSearch, cmd = m.TextAreaSearch.Update(msg)
		cmds = append(cmds, cmd)

		newValue := m.TextAreaSearch.Value()
		if newValue != oldValue && newValue != m.FullSearchQuery {
			m.FullSearchQuery = newValue
			cmds = append(cmds, debouncerFullSearchNote(m, 500*time.Millisecond))
		}
	}

	m.ListModel, cmd = m.ListModel.Update(msg)
	cmds = append(cmds, cmd)

	if selected := m.ListModel.SelectedItem(); selected != nil {
		if note, ok := selected.(noteItem); ok {
			wrapped := wordwrap.String(fmt.Sprintf("%v", note.NoteText), m.TextareaEdit.Width())
			// Só atualize o valor se for diferente do atual
			if m.TextareaEdit.Value() != wrapped {
				m.TextareaEdit.SetValue(wrapped)
			}
		}
	} else {
		if m.TextareaEdit.Value() != "" {
			m.TextareaEdit.SetValue("")
		}
	}

	m.TextareaEdit, cmd = m.TextareaEdit.Update(tea.KeyMsg{Type: tea.KeyNull})
	cmds = append(cmds, cmd)

	return *m, tea.Batch(cmds...)
}

func debouncerFullSearchNote(m *model.Model, d time.Duration) tea.Cmd {
	if m.FullSearchTimerCancel != nil {
		close(m.FullSearchTimerCancel)
	}
	cancel := make(chan struct{})
	m.FullSearchTimerCancel = cancel
	return func() tea.Msg {
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-t.C:
			return fullSearchDebounceMsg{}
		case <-cancel:
			return nil
		}
	}
}

func FullSearchQueryMapNotes(m *model.Model) []list.Item {
	mapQuery, err := m.DB.FullSearchNote(m.Context, m.FullSearchQuery)
	if err != nil {
		return []list.Item{}
	}
	m.MapNotes = mapQuery

	var ids []int
	for id := range m.MapNotes {
		ids = append(ids, id)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))

	items := make([]list.Item, 0, len(mapQuery))
	for _, id := range ids {
		note := m.MapNotes[id]
		noteTimestamp := time.Unix(note.Hour, 0)
		items = append(items, noteItem{
			title:        titleFormatter(note.NoteText),
			desc:         fmt.Sprintf("%v/%d/%v %v:%02d", noteTimestamp.Day(), noteTimestamp.Month(), noteTimestamp.Year(), noteTimestamp.Hour(), noteTimestamp.Minute()),
			NoteText:     note.NoteText,
			Id:           id,
			Reminder:     0,
			PlusReminder: 0,
		})
	}
	return items
}

func getPaginationInfo(m *model.Model) (totalPages int, hasNextPage bool, hasPrevPage bool) {
	totalRows, _ := m.DB.GetTotalCount(m.Context)
	totalPages = (totalRows + PageSize - 1) / PageSize
	if totalPages == 0 {
		totalPages = 1
	}
	hasNextPage = m.CurrentPage < totalPages
	hasPrevPage = m.CurrentPage > 1

	return totalPages, hasNextPage, hasPrevPage
}

func queryMapNotes(m *model.Model) []list.Item {
	mapQuery, err := m.DB.QueryNote(PageSize, (m.CurrentPage-1)*PageSize, m.Context)
	if err != nil {
		file.WriteLog(err.Error(), m.LogPath)
	}
	m.MapNotes = mapQuery

	var ids []int
	for id := range m.MapNotes {
		ids = append(ids, id)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))

	items := make([]list.Item, 0, len(mapQuery))
	for _, id := range ids {
		note := m.MapNotes[id]
		noteTimestamp := time.Unix(note.Hour, 0)
		items = append(items, noteItem{
			title:        titleFormatter(note.NoteText),
			desc:         fmt.Sprintf("%v/%d/%v %v:%02d", noteTimestamp.Day(), noteTimestamp.Month(), noteTimestamp.Year(), noteTimestamp.Hour(), noteTimestamp.Minute()),
			NoteText:     note.NoteText,
			Id:           id,
			Reminder:     0,
			PlusReminder: 0,
		})
	}
	return items
}
