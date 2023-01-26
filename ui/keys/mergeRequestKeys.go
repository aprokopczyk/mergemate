package keys

import "github.com/charmbracelet/bubbles/key"

type MergeRequestKeyMap struct {
	Rebase key.Binding
	Merge  key.Binding
}

func MergeRequestHelp() MergeRequestKeyMap {
	return MergeRequestKeyMap{
		Rebase: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "Rebase")),
		Merge:  key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "Merge")),
	}
}
