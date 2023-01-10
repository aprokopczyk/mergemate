package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyMergeRequest = "mergeRequest"
	columnKeyStatus       = "status"
	columnKeySourceBranch = "sourceBranch"
	columnKeyTargetBranch = "targetBranch"
)

type MergeRequestTable struct {
	flexTable    table.Model
	totalMargin  int
	totalWidth   int
	gitlabClient *gitlab.ApiClient
}

func NewMergeRequestTable(apiClient *gitlab.ApiClient, totalMargin int) *MergeRequestTable {
	return &MergeRequestTable{
		flexTable: table.New([]table.Column{
			table.NewFlexColumn(columnKeyMergeRequest, "Merge request", 1),
			table.NewFlexColumn(columnKeyStatus, "Status", 1),
			table.NewFlexColumn(columnKeySourceBranch, "Source branch", 1),
			table.NewFlexColumn(columnKeyTargetBranch, "Target branch", 1),
		}).WithRows([]table.Row{}).Focused(true).
			HeaderStyle(lipgloss.NewStyle().Bold(true)).
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)),
		gitlabClient: apiClient,
		totalMargin:  totalMargin,
	}
}

func (m *MergeRequestTable) listMergeRequests() tea.Msg {
	mergeRequests := m.gitlabClient.ListMergeRequests()

	return mergeRequests
}

func (m *MergeRequestTable) Init() tea.Cmd {
	return m.listMergeRequests
}

func (m *MergeRequestTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.flexTable, cmd = m.flexTable.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case []gitlab.MergeRequestDetails:
		rows := make([]table.Row, len(msg))
		for i := 0; i < len(msg); i++ {
			rows[i] = table.NewRow(table.RowData{
				columnKeyMergeRequest: msg[i].Title,
				columnKeyStatus:       msg[i].DetailedMergeStatus,
				columnKeySourceBranch: msg[i].SourceBranch,
				columnKeyTargetBranch: msg[i].TargetBranch,
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

func (m *MergeRequestTable) recalculateTable() {
	m.flexTable = m.flexTable.WithTargetWidth(m.totalWidth - m.totalMargin)
}

func (m *MergeRequestTable) View() string {
	return m.flexTable.View() + "\n"
}
