package context

import "github.com/aprokopczyk/mergemate/ui/styles"

type AppContext struct {
	MainContentHeight int
	HelpHeight        int
	WindowWidth       int
	WindowHeight      int
	Styles            styles.Styles
}

type UpdatedContextMessage struct {
}
