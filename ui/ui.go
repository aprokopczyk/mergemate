package ui

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	"github.com/aprokopczyk/mergemate/ui/tabs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const (
	mergeRequestsTab = iota
	branchesTab
	lastTab
)

type UI struct {
	tabs         []string
	tabContent   []tea.Model
	activeTab    int
	gitlabClient *gitlab.ApiClient
	totalWidth   int
}

func New(apiClient *gitlab.ApiClient) *UI {
	ui := &UI{
		tabs:         make([]string, lastTab),
		tabContent:   make([]tea.Model, lastTab),
		activeTab:    mergeRequestsTab,
		gitlabClient: apiClient,
		totalWidth:   0,
	}

	return ui
}

func (ui *UI) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	ui.tabs[mergeRequestsTab] = "Merge requests"
	ui.tabContent[mergeRequestsTab] = tabs.NewMergeRequestTable(ui.gitlabClient, tabContentsStyle.GetHorizontalFrameSize())
	ui.tabs[branchesTab] = "Your branches"
	ui.tabContent[branchesTab] = tabs.NewBranchTable(ui.gitlabClient, tabContentsStyle.GetHorizontalFrameSize())
	cmds = append(cmds, ui.tabContent[mergeRequestsTab].Init())
	cmds = append(cmds, ui.tabContent[branchesTab].Init())
	return tea.Batch(cmds...)
}

func (ui *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tabs.MergeAutomaticallyStatus:
		componentModel, componentCommand := ui.tabContent[mergeRequestsTab].Update(msg)
		ui.tabContent[mergeRequestsTab] = componentModel
		if componentCommand != nil {
			cmds = append(cmds, componentCommand)
		}
		break
	default:
		componentModel, componentCommand := ui.tabContent[ui.activeTab].Update(msg)
		ui.tabContent[ui.activeTab] = componentModel
		if componentCommand != nil {
			cmds = append(cmds, componentCommand)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "right":
			ui.activeTab = min(ui.activeTab+1, len(ui.tabs)-1)
			cmds = append(cmds, ui.tabContent[ui.activeTab].Init())
		case "left":
			ui.activeTab = max(ui.activeTab-1, 0)
			cmds = append(cmds, ui.tabContent[ui.activeTab].Init())
		case "ctrl+c":
			cmds = append(cmds, tea.Quit)
		}
	case tea.WindowSizeMsg:
		ui.totalWidth = msg.Width
		cmds = append(cmds, triggerOnAll(msg, ui)...)
	}

	return ui, tea.Batch(cmds...)
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

var (
	tabStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			Padding(0, 1, 0, 1).
			BorderForeground(colors.Emerald800)
	tabSectionStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, true, false).
			Padding(1, 0, 0, 2).
			BorderForeground(colors.Emerald800)
	tabContentsStyle = lipgloss.NewStyle().Padding(0, 0, 0, 2)
)

func (ui *UI) View() string {
	toRender := strings.Builder{}

	var renderedTabs []string

	for i, t := range ui.tabs {
		var style = tabStyle.Copy()
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
	toRender.WriteString(tabSectionStyle.Copy().Width(ui.totalWidth).Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)))
	toRender.WriteString("\n")
	toRender.WriteString(tabContentsStyle.Copy().Width(ui.totalWidth).Render(ui.tabContent[ui.activeTab].View()))
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
