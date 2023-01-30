package context

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/styles"
)

type AppContext struct {
	MainContentHeight int
	HelpHeight        int
	WindowWidth       int
	WindowHeight      int
	MergeJobInterval  int
	Styles            styles.Styles
	GitlabClient      *gitlab.ApiClient
}

type UpdatedContextMessage struct {
}
