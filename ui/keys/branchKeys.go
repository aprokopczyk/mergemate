package keys

import "github.com/charmbracelet/bubbles/key"

type BranchKeyMap struct {
	MergeAutomatically      key.Binding
	CloseTargetBranchesList key.Binding
	SelectTargetBranch      key.Binding
}

func BranchHelp() BranchKeyMap {
	return BranchKeyMap{
		MergeAutomatically:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "Create automatic merge request")),
		CloseTargetBranchesList: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Close target branches list"), key.WithDisabled()),
		SelectTargetBranch:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Select target branch"), key.WithDisabled()),
	}
}
