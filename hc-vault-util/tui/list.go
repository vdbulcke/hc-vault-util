package tui

import (
	"fmt"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	vault "github.com/hashicorp/vault/api"
)

var (
	UIState *State

	docStyle = lipgloss.NewStyle().Margin(1, 2)
	// WindowSize store the size of the terminal window
	WindowSize tea.WindowSizeMsg
)

type State struct {
	// list of item from list view
	List []list.Item
	// current index from the list for
	// viewport view
	DisplayCurrentIndex int

	// Mount path for kv2
	// secret engine
	Mount string

	// curent secret dir
	Current string

	Client *vault.Client
}

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	insertItem       key.Binding
	viewItem         key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{

		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
	}
}

type model struct {
	current      string
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
	quitting     bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		WindowSize = msg
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleSpinner):
			cmd := m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.delegateKeys.choose):
			// fmt.Println("key selected")
			item, ok := m.list.SelectedItem().(ListItem)
			if ok {
				// set current index in state
				UIState.DisplayCurrentIndex = m.list.Index()

				if item.isSecret() {
					entry, err := newSecretDisplay(item.path)
					if err != nil {
						m.quitting = true
						return m, tea.Quit
					}
					return entry.Update(msg)
				} else {

					p := path.Join(UIState.Current, item.path)

					list, err := GenerateListItemList(UIState.Client, UIState.Mount, p)
					if err != nil {

						m.quitting = true
						return m, tea.Quit
					}

					// update current dir
					UIState.Current = p

					UIState.List = list
					UIState.DisplayCurrentIndex = 0

					resetM, resetCmd := InitList(list, UIState.Current)
					cmds = append(cmds, resetCmd)

					return resetM, tea.Batch(cmds...)
				}

			}

			return m, nil

		case key.Matches(msg, m.delegateKeys.h):

			current := strings.TrimSuffix(UIState.Current, "/")

			parent := path.Dir(current)

			list, err := GenerateListItemList(UIState.Client, UIState.Mount, parent)
			if err != nil {

				m.quitting = true
				return m, tea.Quit
			}

			UIState.Current = parent
			UIState.List = list
			UIState.DisplayCurrentIndex = 0

			resetM, resetCmd := InitList(list, UIState.Current)
			cmds = append(cmds, resetCmd)

			return resetM, tea.Batch(cmds...)

		}

		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {

	return docStyle.Render(m.list.View())
}

func InitList(items []list.Item, current string) (tea.Model, tea.Cmd) {

	listKeys := newListKeyMap()
	delegateKeys := newDelegateKeyMap()
	keyDelegate := newItemDelegate(delegateKeys)
	m := model{
		list:         list.New(items, keyDelegate, 8, 8),
		keys:         listKeys,
		delegateKeys: delegateKeys,
		current:      current,
	}
	m.list.Title = fmt.Sprintf("Kv2: %s/", current)
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleSpinner,
			listKeys.viewItem,
			listKeys.insertItem,
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
		}
	}
	m.list.KeyMap = CustomKeyMap()
	if WindowSize.Height != 0 {
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(WindowSize.Width-left-right, WindowSize.Height-top-bottom-1)
	}

	return m, nil
}

// CustomKeyMap returns a default set of keybindings.
func CustomKeyMap() list.KeyMap {
	return list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "pgup", "b", "u"),
			key.WithHelp("←/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "pgdown", "f", "d"),
			key.WithHelp("→/pgdn", "next page"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),

		// Filtering.
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
		),

		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
}
