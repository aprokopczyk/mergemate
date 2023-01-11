package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"strings"
)

const (
	columnKeyBranchName = "branchName"
	columnKeyLastCommit = "lastCommit"
)

type BranchTable struct {
	branchesList     list.Model
	flexTable        table.Model
	totalMargin      int
	totalWidth       int
	gitlabClient     *gitlab.ApiClient
	branches         []gitlab.Branch
	showMergeTargets bool
}

type branchItem struct {
	name string
}

func (i branchItem) Title() string       { return i.name }
func (i branchItem) Description() string { return i.name }
func (i branchItem) FilterValue() string { return i.name }

func NewBranchTable(apiClient *gitlab.ApiClient, totalMargin int) *BranchTable {
	return &BranchTable{
		flexTable: table.New([]table.Column{
			table.NewFlexColumn(columnKeyBranchName, "Branch", 15),
			table.NewFlexColumn(columnKeyLastCommit, "Last commit date", 15),
		}).WithRows([]table.Row{}).
			Focused(true).
			HeaderStyle(lipgloss.NewStyle().Bold(true)).
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)),
		gitlabClient:     apiClient,
		totalMargin:      totalMargin,
		branches:         []gitlab.Branch{},
		showMergeTargets: false,
		branchesList:     createList(),
	}
}

func createList() list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	model := list.New([]list.Item{}, delegate, 0, 20)
	model.Title = "Select target branch"
	model.DisableQuitKeybindings()
	model.SetShowStatusBar(false)
	return model
}

func (m *BranchTable) listBranches() tea.Msg {
	branches := m.gitlabClient.ListBranches()

	return branches
}

func (m *BranchTable) createMergeRequest(sourceBranch string, targetBranch string, title string) tea.Cmd {
	return func() tea.Msg {
		m.gitlabClient.CreateMergeRequest(sourceBranch, targetBranch, title)
		return nil
	}
}

func (m *BranchTable) Init() tea.Cmd {
	return m.listBranches
}

func (m *BranchTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if !m.showMergeTargets {
		m.flexTable, cmd = m.flexTable.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.branchesList, cmd = m.branchesList.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case []gitlab.Branch:
		var rows []table.Row
		var targetBranches []list.Item
		for i := 0; i < len(msg); i++ {
			if strings.HasPrefix(msg[i].Name, m.gitlabClient.BranchPrefix) {
				rows = append(rows, table.NewRow(table.RowData{
					columnKeyBranchName: msg[i].Name,
					columnKeyLastCommit: msg[i].Commit.AuthoredDate,
				}))
			} else {
				targetBranches = append(targetBranches, branchItem{name: msg[i].Name})
			}
		}
		m.flexTable = m.flexTable.WithRows(rows)
		m.branchesList.SetItems(targetBranches)
		m.branches = msg
	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.recalculateComponents()
		cmds = append(cmds, tea.ClearScreen)
	case tea.KeyMsg:
		switch msg.String() {
		case "m":
			if !m.showMergeTargets {
				m.showMergeTargets = true
				m.recalculateComponents()
			}
		case "esc":
			if m.showMergeTargets {
				m.showMergeTargets = false
				m.recalculateComponents()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *BranchTable) recalculateComponents() {
	tableWidth := m.tableSize()
	m.flexTable = m.flexTable.WithTargetWidth(tableWidth)
	v := m.contentSize() - tableWidth
	m.branchesList.SetWidth(v)
}

func (m *BranchTable) tableSize() int {
	contentSize := m.contentSize()
	if m.showMergeTargets {
		return int(float64(contentSize) * 0.7)
	}
	return contentSize
}

func (m *BranchTable) contentSize() int {
	var contentSize = m.totalWidth - m.totalMargin
	return contentSize
}

func (m *BranchTable) View() string {
	if m.showMergeTargets {
		return lipgloss.JoinHorizontal(lipgloss.Top, m.flexTable.View(), m.branchesList.View())
	}
	return m.flexTable.View() + "\n"
}