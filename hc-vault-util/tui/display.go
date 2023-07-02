package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	vault "github.com/hashicorp/vault/api"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type SecretDisplay struct {
	viewport   viewport.Model
	wasApplied bool
	path       string
}

func newSecretDisplay(path string) (*SecretDisplay, error) {
	// const width = 78

	top, right, bottom, left := lipgloss.NewStyle().Margin(0, 2).GetMargin()
	vp := viewport.New(WindowSize.Width-left-right, WindowSize.Height-top-bottom-6)
	s := &SecretDisplay{
		viewport:   vp,
		wasApplied: false,
		path:       path,
	}
	s.viewport.Style = lipgloss.NewStyle().Align(lipgloss.Bottom)

	str, _ := glamour.Render(s.genMDPatchInfo(UIState), "dark")
	s.viewport.SetContent(str)

	return s, nil
}
func (s SecretDisplay) Init() tea.Cmd {
	return nil
}

func (s SecretDisplay) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		WindowSize = msg
		top, right, bottom, left := lipgloss.NewStyle().Margin(0, 2).GetMargin()
		s.viewport = viewport.New(WindowSize.Width-left-right, WindowSize.Height-top-bottom-6)

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c":
			return s, tea.Quit
		case "q", "esc", "h":

			return InitList(UIState.List, UIState.Current)

		default:
			var cmd tea.Cmd
			s.viewport, cmd = s.viewport.Update(msg)
			return s, cmd
		}
	default:
		return s, nil
	}
	return s, nil
}

func (s SecretDisplay) View() string {
	return s.viewport.View() + s.helpView()
}

func (s SecretDisplay) helpView() string {

	return helpStyle("\n  ↑/↓: Navigate • q|esc|h: Back • ctrl+C: Quit\n")
}

func (s *SecretDisplay) getVaultSecret(client *vault.Client, mount, path string) (*vault.KVSecret, error) {
	ctx := context.Background()
	secret, err := client.KVv2(mount).Get(ctx, path)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (s SecretDisplay) genMDPatchInfo(state *State) string {

	p := path.Join(UIState.Current, s.path)
	kv, err := s.getVaultSecret(state.Client, state.Mount, p)
	if err != nil {
		return genMDError(err)
	}

	dataBytes, err := json.MarshalIndent(kv.Data, "", "  ")
	if err != nil {
		return genMDError(err)
	}
	metadataBytes, err := json.MarshalIndent(kv.CustomMetadata, "", "  ")
	if err != nil {
		return genMDError(err)
	}

	mdTemp := `
# Current Secret 

* Path: %s
* Version: %d

	
## Data
%s
%s
%s

## Version 
	
* CreatedTime: %s
* DeletedTime: %s

## Metadata

%s
%s
%s

`
	out := fmt.Sprintf(mdTemp,
		p,
		kv.VersionMetadata.Version,
		"```json",
		string(dataBytes),
		"```",
		kv.VersionMetadata.CreatedTime.String(),
		kv.VersionMetadata.DeletionTime.String(),
		"```json",
		string(metadataBytes),
		"```",
	)

	return out

}

func genMDError(err error) string {
	errTemplate := `
# Error 

%s
`
	return fmt.Sprintf(errTemplate, err.Error())
}
