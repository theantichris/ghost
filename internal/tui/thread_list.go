package tui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/theantichris/ghost/v3/internal/storage"
)

// ThreadListModel holds the state for the thread list.
type ThreadListModel struct {
	list          list.Model
	store         *storage.Store
	width, height int
	logger        *log.Logger
}

// NewThreadListModel creates a new model and stores the current list of threads.
func NewThreadListModel(store *storage.Store, width, height int, logger *log.Logger) (ThreadListModel, error) {
	threads, err := store.ListThreads()
	if err != nil {
		return ThreadListModel{}, err
	}

	// Convert each thread into a listItems item.
	var listItems []list.Item
	for _, thread := range threads {
		listItems = append(listItems, threadItem{thread: thread})
	}

	list := list.New(listItems, list.NewDefaultDelegate(), width, height)

	list.Title = "Threads"

	model := ThreadListModel{
		list:   list,
		store:  store,
		width:  width,
		height: height,
		logger: logger,
	}

	return model, nil
}

// Init handles initializing model state, current does nothing.
func (model ThreadListModel) Init() tea.Cmd {
	return nil
}

// Update updates model state.
func (model ThreadListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	model.list, cmd = model.list.Update(msg)

	return model, cmd
}

// View renders model state.
func (model ThreadListModel) View() tea.View {
	view := tea.NewView(
		lipgloss.Place(model.width, model.height, lipgloss.Left, lipgloss.Center, model.list.View()),
	)

	return view
}
