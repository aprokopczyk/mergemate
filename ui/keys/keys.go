package keys

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up                    key.Binding
	Down                  key.Binding
	Left                  key.Binding
	Right                 key.Binding
	Quit                  key.Binding
	ActiveContextBindings []key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Quit},
		k.ActiveContextBindings,
	}
}

func GetKeyMap(activeContextBindings []key.Binding) help.KeyMap {
	keys := Keys
	keys.ActiveContextBindings = activeContextBindings
	return keys
}

var Keys = keyMap{
	Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Left:  key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "Switch to left tab")),
	Right: key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "Switch to right tab")),
	Quit:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Quit")),
}
