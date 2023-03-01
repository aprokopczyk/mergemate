package context

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/styles"
)

type AppContext struct {
	TableContentHeight   int
	HelpHeight           int
	WindowWidth          int
	WindowHeight         int
	TablePageSize        int
	MergeJobInterval     int
	Styles               styles.Styles
	GitlabClient         *gitlab.ApiClient
	UserBranchPrefix     string
	TargetBranchPrefixes []string
	FavouriteBranches    []string
}

type UpdatedContextMessage struct {
}
