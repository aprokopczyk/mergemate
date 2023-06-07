package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	"github.com/aprokopczyk/mergemate/ui/context"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"log"
	"strings"
)

type MergedMergeRequestTable struct {
	flexTable     table.Model
	mergeRequests []gitlab.MergeRequestDetails
	context       *context.AppContext
}

func NewMergedMergeRequestTable(context *context.AppContext) *MergedMergeRequestTable {
	return &MergedMergeRequestTable{
		flexTable: table.New([]table.Column{
			table.NewFlexColumn(columnKeyMergeRequest, "Merge request", 1),
			table.NewFlexColumn(columnKeySourceBranch, "Source branch", 1),
			table.NewFlexColumn(columnKeyTargetBranch, "Target branch", 1),
		}).WithRows([]table.Row{}).Focused(true).
			HeaderStyle(lipgloss.NewStyle().Bold(true)).
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)).
			WithPageSize(context.TablePageSize),
		context: context,
	}
}

func (m *MergedMergeRequestTable) listMergeRequests() tea.Msg {
	mergeRequests, err := m.context.GitlabClient.MergedMergeRequests()
	if err != nil {
		log.Printf("Error when fetching merge requests %v", err)
	}
	return mergeRequests
}

func (m *MergedMergeRequestTable) shouldBeMergedAutomatically(mergeRequestIid int) tea.Cmd {
	return func() tea.Msg {
		notes, err := m.context.GitlabClient.ListMergeRequestNotes(mergeRequestIid)
		if err != nil {
			log.Printf("Error when fetching merge request notes %v", err)
		}
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

func (m *MergedMergeRequestTable) Init() tea.Cmd {
	return m.listMergeRequests
}

func (m *MergedMergeRequestTable) Update(msg tea.Msg) (TabContent, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.flexTable, cmd = m.flexTable.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case []gitlab.MergeRequestDetails:
		m.mergeRequests = msg
		m.redrawTable()
		m.flexTable = m.flexTable.PageFirst()
	case context.UpdatedContextMessage:
		m.recalculateTable()
	}

	return m, tea.Batch(cmds...)
}

func (m *MergedMergeRequestTable) redrawTable() {
	var rows []table.Row
	for _, mergeRequest := range m.mergeRequests {
		rows = append(rows, table.NewRow(table.RowData{
			columnKeyMergeRequest:         mergeRequest.Title,
			columnKeySourceBranch:         mergeRequest.SourceBranch,
			columnKeyTargetBranch:         mergeRequest.TargetBranch,
			columnKeyMergeRequestMetadata: mergeRequest,
		}))
	}
	m.flexTable = m.flexTable.WithRows(rows)
}

func (m *MergedMergeRequestTable) recalculateTable() {
	m.flexTable = m.flexTable.WithTargetWidth(m.context.WindowWidth - m.context.Styles.Tabs.Content.GetHorizontalFrameSize())
	m.flexTable = m.flexTable.WithPageSize(m.context.TablePageSize)
}

func (m *MergedMergeRequestTable) FullHelp() []key.Binding {
	return []key.Binding{}
}

func (m *MergedMergeRequestTable) View() string {
	return m.flexTable.View()
}
