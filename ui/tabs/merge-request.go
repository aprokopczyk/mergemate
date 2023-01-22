package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"strings"
)

const (
	columnKeyMergeRequest       = "mergeRequest"
	columnKeyMergeAutomatically = "mergeAutomatically"
	columnKeyStatus             = "status"
	columnKeySourceBranch       = "sourceBranch"
	columnKeyTargetBranch       = "targetBranch"
)

const checking = "checking"
const yes = "yes"
const no = "no"

const MergeAutomatically = "MERGE_AUTOMATICALLY"

type MergeRequestWithMetadata struct {
	mergeRequest         gitlab.MergeRequestDetails
	automaticMergeStatus string
}

type MergeRequestTable struct {
	flexTable                  table.Model
	mergeAutomaticallyStatuses map[int]string
	mergeRequests              []gitlab.MergeRequestDetails
	totalMargin                int
	totalWidth                 int
	gitlabClient               *gitlab.ApiClient
}

func NewMergeRequestTable(apiClient *gitlab.ApiClient, totalMargin int) *MergeRequestTable {
	return &MergeRequestTable{
		flexTable: table.New([]table.Column{
			table.NewFlexColumn(columnKeyMergeRequest, "Merge request", 1),
			table.NewFlexColumn(columnKeyMergeAutomatically, "Merge automatically", 1),
			table.NewFlexColumn(columnKeyStatus, "Status", 1),
			table.NewFlexColumn(columnKeySourceBranch, "Source branch", 1),
			table.NewFlexColumn(columnKeyTargetBranch, "Target branch", 1),
		}).WithRows([]table.Row{}).Focused(true).
			HeaderStyle(lipgloss.NewStyle().Bold(true)).
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)),
		gitlabClient:               apiClient,
		totalMargin:                totalMargin,
		mergeAutomaticallyStatuses: make(map[int]string),
	}
}

func (m *MergeRequestTable) listMergeRequests() tea.Msg {
	mergeRequests := m.gitlabClient.ListMergeRequests()

	return mergeRequests
}

type MergeAutomaticallyStatus struct {
	mergeRequestIid             int
	shouldBeMergedAutomatically bool
}

func (m *MergeRequestTable) shouldBeMergedAutomatically(mergeRequestIid int) tea.Cmd {
	return func() tea.Msg {
		notes := m.gitlabClient.ListMergeRequestNotes(mergeRequestIid)
		var shouldBeMergedAutomatically bool
		for _, note := range notes {
			if strings.HasPrefix(note.Body, MergeAutomatically) {
				shouldBeMergedAutomatically = true
				break
			}
		}
		return MergeAutomaticallyStatus{
			mergeRequestIid:             mergeRequestIid,
			shouldBeMergedAutomatically: shouldBeMergedAutomatically,
		}
	}
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
		mergeAutomaticallyStatuses := make(map[int]string)
		for i := 0; i < len(msg); i++ {
			mrIid := msg[i].Iid
			oldEntry, exists := m.mergeAutomaticallyStatuses[mrIid]
			var mergeAutomaticallyStatus = checking
			if exists {
				mergeAutomaticallyStatus = oldEntry
			} else {
				cmds = append(cmds, m.shouldBeMergedAutomatically(mrIid))
			}
			mergeAutomaticallyStatuses[mrIid] = mergeAutomaticallyStatus
		}
		m.mergeAutomaticallyStatuses = mergeAutomaticallyStatuses
		m.mergeRequests = msg
		m.redrawTable()
	case MergeAutomaticallyStatus:
		shouldBeMerged := no
		if msg.shouldBeMergedAutomatically {
			shouldBeMerged = yes
		}
		m.mergeAutomaticallyStatuses[msg.mergeRequestIid] = shouldBeMerged
		m.redrawTable()
	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.recalculateTable()
		cmds = append(cmds, tea.ClearScreen)
	}

	return m, tea.Batch(cmds...)
}

func (m *MergeRequestTable) redrawTable() {
	var rows []table.Row
	for _, mergeRequest := range m.mergeRequests {
		rows = append(rows, table.NewRow(table.RowData{
			columnKeyMergeRequest:       mergeRequest.Title,
			columnKeyMergeAutomatically: m.mergeAutomaticallyStatuses[mergeRequest.Iid],
			columnKeyStatus:             mergeRequest.DetailedMergeStatus,
			columnKeySourceBranch:       mergeRequest.SourceBranch,
			columnKeyTargetBranch:       mergeRequest.TargetBranch,
		}))
	}
	m.flexTable = m.flexTable.WithRows(rows)
}

func (m *MergeRequestTable) recalculateTable() {
	m.flexTable = m.flexTable.WithTargetWidth(m.totalWidth - m.totalMargin)
}

func (m *MergeRequestTable) View() string {
	return m.flexTable.View() + "\n"
}
