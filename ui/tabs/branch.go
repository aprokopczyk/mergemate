package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyBranchName = "branchName"
	columnKeyLastCommit = "lastCommit"
)

type BranchTable struct {
	flexTable    table.Model
	totalMargin  int
	totalWidth   int
	gitlabClient *gitlab.ApiClient
}

func NewBranchTable(apiClient *gitlab.ApiClient, totalMargin int) *BranchTable {
	return &BranchTable{
		flexTable: table.New([]table.Column{
			table.NewFlexColumn(columnKeyBranchName, "Branch", 15),
			table.NewFlexColumn(columnKeyLastCommit, "Last commit date", 15),
		}).WithRows([]table.Row{}).
			Focused(true).
			HeaderStyle(lipgloss.NewStyle().Bold(true)).
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)),
		gitlabClient: apiClient,
		totalMargin:  totalMargin,
	}
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

	m.flexTable, cmd = m.flexTable.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case []gitlab.Branch:
		rows := make([]table.Row, len(msg))
		for i := 0; i < len(msg); i++ {
			rows[i] = table.NewRow(table.RowData{
				columnKeyBranchName: msg[i].Name,
				columnKeyLastCommit: msg[i].Commit.AuthoredDate,
			})
		}
		m.flexTable = m.flexTable.WithRows(rows)
	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.recalculateTable()
		cmds = append(cmds, tea.ClearScreen)
	}

	return m, tea.Batch(cmds...)
}

func (m *BranchTable) recalculateTable() {
	m.flexTable = m.flexTable.WithTargetWidth(m.totalWidth - 2)
}

func (m *BranchTable) View() string {
	return m.flexTable.View() + "\n"
}
