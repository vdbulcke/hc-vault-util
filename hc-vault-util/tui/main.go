package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	vault "github.com/hashicorp/vault/api"
)

type ListItem struct {
	path string
}

// implement list item interface for UI
func (p ListItem) Title() string {

	return p.path
}

func (p ListItem) Description() string {

	if p.isSecret() {
		return "Type: secret"
	}
	return "Type: dir"
}

func (p ListItem) isSecret() bool {
	// return path.IsAbs(p.path)
	return !strings.HasSuffix(p.path, "/")
}

func (p ListItem) FilterValue() string { return p.path }

func vaultListPath(client *vault.Client, mount, current string) ([]string, error) {

	// NOTE: hack cannot list kv2 secrets directly
	//       but can list secret metadata
	path := fmt.Sprintf("%s/metadata/%s", mount, current)
	secret, err := client.Logical().List(path)
	if err != nil {
		return nil, err
	}
	keys, ok := secret.Data["keys"].([]interface{})
	if ok {

		l := make([]string, len(keys))
		for i, k := range keys {

			l[i] = k.(string)

		}

		return l, nil
	}

	return []string{}, nil

}

// GenerateListItemList filter paths and format them to ListItem
func GenerateListItemList(client *vault.Client, mount, current string) ([]list.Item, error) {

	paths, err := vaultListPath(client, mount, current)
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(paths))
	for i, p := range paths {
		item := ListItem{
			path: p,
		}
		items[i] = list.Item(item)

	}

	return items, nil

}

func GenerateItemList(pitems []ListItem) []list.Item {
	items := make([]list.Item, len(pitems))
	for i, p := range pitems {

		items[i] = list.Item(p)

	}

	return items
}

func StartUI(client *vault.Client, mount string) error {

	items, err := GenerateListItemList(client, mount, "")
	if err != nil {
		return err
	}

	UIState = &State{
		List:                items,
		DisplayCurrentIndex: 0,
		Client:              client,
		Mount:               mount,
	}
	m, _ := InitList(UIState.List, "")
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
