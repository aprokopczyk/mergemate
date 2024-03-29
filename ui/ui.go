package ui

import (
	"github.com/aprokopczyk/mergemate/ui/colors"
	"github.com/aprokopczyk/mergemate/ui/context"
	"github.com/aprokopczyk/mergemate/ui/keys"
	"github.com/aprokopczyk/mergemate/ui/styles"
	"github.com/aprokopczyk/mergemate/ui/tabs"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const (
	activeMergeRequestsTab = iota
	branchesTab
	mergedMergeRequestsTab
	lastTab
)

type UI struct {
	tabs       []string
	tabContent []tabs.TabContent
	activeTab  int
	actionLog  *tabs.ActionLog
	help       help.Model
	context    *context.AppContext
}

func New(context *context.AppContext) *UI {
	helpModel := help.New()
	helpModel.ShowAll = true
	ui := &UI{
		tabs:       make([]string, lastTab),
		tabContent: make([]tabs.TabContent, lastTab),
		activeTab:  activeMergeRequestsTab,
		actionLog:  tabs.NewActionLog(context),
		help:       helpModel,
		context:    context,
	}

	return ui
}

func (ui *UI) listTargetBranches() tea.Msg {
	branches := ui.context.GitlabClient.FetchBranchesWithPattern(ui.context.TargetBranchPrefixes)
	return tabs.TargetBranches{Branches: branches}
}

func (ui *UI) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	ui.tabs[activeMergeRequestsTab] = "Active merge requests"
	ui.tabContent[activeMergeRequestsTab] = tabs.NewActiveMergeRequestTable(ui.context)
	ui.tabs[branchesTab] = "Your branches"
	ui.tabContent[branchesTab] = tabs.NewBranchTable(ui.context)
	ui.tabs[mergedMergeRequestsTab] = "Merged merge requests"
	ui.tabContent[mergedMergeRequestsTab] = tabs.NewMergedMergeRequestTable(ui.context)
	cmds = append(cmds, ui.tabContent[activeMergeRequestsTab].Init())
	cmds = append(cmds, ui.tabContent[branchesTab].Init())
	cmds = append(cmds, ui.tabContent[mergedMergeRequestsTab].Init())
	cmds = append(cmds, ui.listTargetBranches)
	return tea.Batch(cmds...)
}

func (ui *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	actionLog, cmd := ui.actionLog.Update(msg)
	ui.actionLog = actionLog
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Right):
			ui.activeTab = min(ui.activeTab+1, len(ui.tabs)-1)
			cmds = append(cmds, ui.tabContent[ui.activeTab].Init())
		case key.Matches(msg, keys.Keys.Left):
			ui.activeTab = max(ui.activeTab-1, 0)
			cmds = append(cmds, ui.tabContent[ui.activeTab].Init())
		case key.Matches(msg, keys.Keys.Quit):
			cmds = append(cmds, tea.Quit)
		}
	case tea.WindowSizeMsg:
		ui.context.WindowHeight = msg.Height
		ui.context.WindowWidth = msg.Width
		ui.context.TableContentHeight = msg.Height - styles.TabsHeaderHeight - ui.getHelpHeight() - ui.actionLog.Height

		newPageSize := ui.context.TableContentHeight - styles.TableHeaderHeight - styles.TableFooterHeight
		if newPageSize > styles.MinTablePageSize {
			ui.context.TablePageSize = newPageSize
		} else {
			ui.context.TablePageSize = styles.MinTablePageSize
		}

		ui.help.Width = msg.Width
		cmds = append(cmds, tea.ClearScreen)
		cmds = append(cmds, triggerOnAll(context.UpdatedContextMessage{}, ui)...)
	}

	// key message only to active tab, rest goes to all tabs
	if _, ok := msg.(tea.KeyMsg); ok {
		componentModel, componentCommand := ui.tabContent[ui.activeTab].Update(msg)
		ui.tabContent[ui.activeTab] = componentModel
		if componentCommand != nil {
			cmds = append(cmds, componentCommand)
		}
	} else {
		cmds = append(cmds, triggerOnAll(msg, ui)...)
	}

	return ui, tea.Batch(cmds...)
}

func (ui *UI) getHelpHeight() int {
	keyMap := keys.GetKeyMap(ui.tabContent[ui.activeTab].FullHelp())
	height := 0
	for _, bindings := range keyMap.FullHelp() {
		height = max(height, len(bindings))
	}
	return height + ui.context.Styles.Help.GetVerticalFrameSize()
}

func triggerOnAll(msg tea.Msg, ui *UI) []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := 0; i < lastTab; i++ {
		componentModel, componentCommand := ui.tabContent[i].Update(msg)
		ui.tabContent[i] = componentModel
		if componentCommand != nil {
			cmds = append(cmds, componentCommand)
		}
	}
	return cmds
}

func (ui *UI) View() string {
	toRender := strings.Builder{}

	var renderedTabs []string

	styleDefinitions := ui.context.Styles
	for i, t := range ui.tabs {
		var style = styleDefinitions.Tabs.TabItem.Copy()
		isActive := i == ui.activeTab
		isLast := i == len(ui.tabs)-1
		if isActive {
			style.Bold(true).Underline(true).Background(colors.Emerald600)
		}
		if isLast {
			style.UnsetBorderRight()
		}

		renderedTabs = append(renderedTabs, style.Render(t))
	}
	toRender.WriteString(styleDefinitions.Tabs.Header.Copy().Width(ui.context.WindowWidth).Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)))
	toRender.WriteString("\n")
	tabsContent := styleDefinitions.Tabs.Content.Copy().Width(ui.context.WindowWidth).Render(ui.tabContent[ui.activeTab].View())
	toRender.WriteString(tabsContent)
	toRender.WriteString("\n")
	// fill up all available space to push footer to the bottom
	toRender.WriteString(strings.Repeat("\n", max(0, ui.context.TableContentHeight-lipgloss.Height(tabsContent))))
	toRender.WriteString(ui.actionLog.View())
	toRender.WriteString("\n")
	toRender.WriteString(styleDefinitions.Help.Copy().Width(ui.context.WindowWidth).Render(ui.help.View(keys.GetKeyMap(ui.tabContent[ui.activeTab].FullHelp()))))
	return toRender.String()
}

func max(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a int, b int) int {
	if a < b {
		return a
	}

	return b
}
