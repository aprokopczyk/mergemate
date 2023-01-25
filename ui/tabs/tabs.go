package tabs

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type TabContent interface {
	Init() tea.Cmd
	Update(tea.Msg) (TabContent, tea.Cmd)
	View() string
	FullHelp() []key.Binding
}
