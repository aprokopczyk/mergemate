package tabs

import (
	"container/ring"
	"fmt"
	"github.com/aprokopczyk/mergemate/ui/context"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	numberOfActions = 5
	headerHeight    = 2
)

type ActionMessage struct {
	Content string
	Success bool
}

func failed(content string) ActionMessage {
	return ActionMessage{
		Content: "Error: " + content,
		Success: false,
	}
}

func success(content string) ActionMessage {
	return ActionMessage{
		Content: "Success: " + content,
		Success: true,
	}
}

type ActionLog struct {
	buffer  *ring.Ring
	context *context.AppContext
	Height  int
}

func NewActionLog(context *context.AppContext) *ActionLog {
	return &ActionLog{
		buffer:  ring.New(numberOfActions),
		context: context,
		Height:  numberOfActions + headerHeight,
	}
}

func (model ActionLog) Init() tea.Cmd {
	return nil
}

func (model ActionLog) Update(msg tea.Msg) (*ActionLog, tea.Cmd) {
	switch msg := msg.(type) {
	case MergeRequestCreated:
		mergeRequest := msg.mergeRequest
		model.buffer.Value = success(fmt.Sprintf("Created merge request: '%s' from branch %s", mergeRequest.Title, mergeRequest.SourceBranch, mergeRequest.Iid))
		model.buffer = model.buffer.Next()
	case ActionMessage:
		model.buffer.Value = msg
		model.buffer = model.buffer.Next()
	}
	return &model, nil
}

func (model ActionLog) View() string {
	appContext := model.context
	styleDefinitions := appContext.Styles
	var messages = ""
	model.buffer.Do(func(a any) {
		lineResult := "\n"
		if a != nil {
			lineResult = lineResult + a.(ActionMessage).Content

		}
		messages = lineResult + messages
	})
	messages = "Recent activity: " + messages
	return styleDefinitions.ActionLog.Copy().Width(appContext.WindowWidth).Render(messages)
}
