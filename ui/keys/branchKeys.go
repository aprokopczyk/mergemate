package keys

import (
	"github.com/charmbracelet/bubbles/key"
	"strconv"
)

type BranchKeyMap struct {
	MergeAutomatically      key.Binding
	CloseTargetBranchesList key.Binding
	SelectTargetBranch      key.Binding
	MergeFavourite          []key.Binding
}

func BranchHelp(branchTargets []string) BranchKeyMap {
	branchKeyMap := BranchKeyMap{
		MergeAutomatically:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "Create automatic merge request")),
		CloseTargetBranchesList: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Close target branches list"), key.WithDisabled()),
		SelectTargetBranch:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Select target branch"), key.WithDisabled()),
	}

	for i, branch := range branchTargets {
		if branch != "" {
			keySymbol := strconv.Itoa(i)
			branchKeyMap.MergeFavourite = append(branchKeyMap.MergeFavourite, key.NewBinding(key.WithKeys(keySymbol), key.WithHelp(keySymbol, "Merge automatically to "+branch)))
		}
	}

	return branchKeyMap
}
