package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	"github.com/aprokopczyk/mergemate/ui/context"
	"github.com/aprokopczyk/mergemate/ui/keys"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"log"
	"strings"
	"time"
)

const (
	columnKeyMergeRequest         = "mergeRequest"
	columnKeyMergeAutomatically   = "mergeAutomatically"
	columnKeyStatus               = "status"
	columnKeySourceBranch         = "sourceBranch"
	columnKeyTargetBranch         = "targetBranch"
	columnKeyMergeRequestMetadata = "mergeRequestMetadata"
)

const checking = "checking"
const yes = "yes"
const no = "no"

const MergeAutomatically = "MERGE_AUTOMATICALLY"

type MergeRequestWithMetadata struct {
	mergeRequest         gitlab.MergeRequestDetails
	automaticMergeStatus string
}

type RequestMetadata struct {
	mergeAutomatically string
	status             string
}

type ActiveMergeRequestTable struct {
	flexTable     table.Model
	mrMetadata    map[int]RequestMetadata
	mergeRequests []gitlab.MergeRequestDetails
	context       *context.AppContext
	keys          keys.MergeRequestKeyMap
}

func NewActiveMergeRequestTable(context *context.AppContext) *ActiveMergeRequestTable {
	return &ActiveMergeRequestTable{
		flexTable: table.New([]table.Column{
			table.NewFlexColumn(columnKeyMergeRequest, "Merge request", 1),
			table.NewFlexColumn(columnKeyMergeAutomatically, "Merge automatically", 1),
			table.NewFlexColumn(columnKeyStatus, "Status", 1),
			table.NewFlexColumn(columnKeySourceBranch, "Source branch", 1),
			table.NewFlexColumn(columnKeyTargetBranch, "Target branch", 1),
		}).WithRows([]table.Row{}).Focused(true).
			HeaderStyle(lipgloss.NewStyle().Bold(true)).
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)).
			WithPageSize(context.TablePageSize),
		context:    context,
		mrMetadata: make(map[int]RequestMetadata),
		keys:       keys.MergeRequestHelp(),
	}
}

func (m *ActiveMergeRequestTable) listMergeRequests() tea.Msg {
	mergeRequests, err := m.context.GitlabClient.OpenedMergeRequests()
	if err != nil {
		log.Printf("Error when fetching merge requests %v", err)
	}
	return mergeRequests
}

type MergeAutomaticallyStatus struct {
	mergeRequestIid             int
	shouldBeMergedAutomatically bool
}

func (m *ActiveMergeRequestTable) shouldBeMergedAutomatically(mergeRequestIid int) tea.Cmd {
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

type MergeRequestProcessingResult struct {
	mrStatus map[int]string
}

const merged = "Merged"

func (m *ActiveMergeRequestTable) processMergeRequests(mergeRequests map[int]bool) tea.Cmd {
	return tea.Tick(time.Second*time.Duration(m.context.MergeJobInterval), func(t time.Time) tea.Msg {
		log.Printf("Processing merge requests: %v", mergeRequests)
		mrStatus := make(map[int]string)
		var rebasing []int

		for mergeRequestIid, shouldBeMerged := range mergeRequests {
			mergeRequest, err := m.context.GitlabClient.GetMergeRequestDetails(mergeRequestIid)
			if err != nil {
				log.Printf("Fetching merge request details failed %v", err)
				continue
			}
			if mergeRequest.RebaseInProgress {
				log.Printf("Merge request {id = %v, title=%v} is being rebased.", mergeRequest.Iid, mergeRequest.Title)
				mrStatus[mergeRequestIid] = "Rebase in progress"
				continue
			} else if mergeRequest.RebaseError != "" && mergeRequest.HasConflicts {
				mrStatus[mergeRequestIid] = "Merge conflict"
				continue
			}
			pipelines, err := m.context.GitlabClient.GetMergeRequestPipelines(mergeRequestIid)
			if err != nil {
				log.Printf("Error when fetching pipeline for merge request{id = %v, title=%v}: %v", mergeRequestIid, mergeRequest.Title, err)
			}
			if gitlab.IsPipelineRunning(pipelines) {
				mrStatus[mergeRequestIid] = "CI running"
				continue
			}
			if len(pipelines) > 0 && pipelines[0].Status == "failed" {
				mrStatus[mergeRequestIid] = "CI failed"
				continue
			}
			isBehindTargetBranch := mergeRequest.CommitsBehind > 0
			if shouldBeMerged && isBehindTargetBranch {
				log.Printf("Merge request {id = %v, title=%v} is behind target branch by %v commits, it will be rebased.", mergeRequestIid, mergeRequest.Title, mergeRequest.CommitsBehind)
				// we will rebase outside loop, in case there is mr that could be merged, we'll need to rebase only once
				rebasing = append(rebasing, mergeRequestIid)
				continue
			} else if isBehindTargetBranch {
				mrStatus[mergeRequestIid] = "Needs rebase"
				continue
			}
			automaticMergeAllowed := gitlab.IsAutomaticMergeAllowed(pipelines)
			if shouldBeMerged && automaticMergeAllowed {
				// hurray, we can merge it!
				log.Printf("Merging merge request {id = %v, title=%v}.", mergeRequestIid, mergeRequest.Title)
				// we pass sha to make sure that nothing was pushed in the meantime
				request, err := m.context.GitlabClient.MergeMergeRequest(mergeRequestIid, mergeRequest.Sha)
				if err != nil {
					log.Printf("Error when merging merge request {id = %v, title=%v}: %v ", mergeRequestIid, mergeRequest.Title, err)
					mrStatus[mergeRequestIid] = "Merge failed"
					continue
				}
				if request.State == "merged" {
					log.Printf("Merged merge request {id = %v, title=%v}.", mergeRequestIid, mergeRequest.Title)
					mrStatus[mergeRequestIid] = merged
				}
				continue
			} else if automaticMergeAllowed {
				mrStatus[mergeRequestIid] = "Ready to merge"
			}
		}

		for _, mrIid := range rebasing {
			mrStatus[mrIid] = "Rebase in progress"
			err := m.context.GitlabClient.RebaseMergeRequest(mrIid, true)
			if err != nil {
				log.Printf("Error when rebasing merge request {id = %v}: %v", mrIid, err)
			}
		}

		return MergeRequestProcessingResult{
			mrStatus: mrStatus,
		}
	})
}

func (m *ActiveMergeRequestTable) Init() tea.Cmd {
	return tea.Batch(m.listMergeRequests, m.processMergeRequests(map[int]bool{}))
}

func (m *ActiveMergeRequestTable) Update(msg tea.Msg) (TabContent, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.flexTable, cmd = m.flexTable.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case []gitlab.MergeRequestDetails:
		mergeAutomaticallyStatuses := make(map[int]RequestMetadata)
		for i := 0; i < len(msg); i++ {
			mrIid := msg[i].Iid
			oldEntry, exists := m.mrMetadata[mrIid]
			var mergeAutomaticallyStatus = RequestMetadata{
				mergeAutomatically: checking,
				status:             checking,
			}
			if exists {
				mergeAutomaticallyStatus = oldEntry
			} else {
				cmds = append(cmds, m.shouldBeMergedAutomatically(mrIid))
			}
			mergeAutomaticallyStatuses[mrIid] = mergeAutomaticallyStatus
		}
		m.mrMetadata = mergeAutomaticallyStatuses
		m.mergeRequests = msg
		m.redrawTable()
		m.flexTable = m.flexTable.PageFirst()
	case MergeRequestCreated:
		cmds = append(cmds, m.listMergeRequests)
	case MergeAutomaticallyStatus:
		shouldBeMerged := no
		if msg.shouldBeMergedAutomatically {
			shouldBeMerged = yes
		}
		metadata, exists := m.mrMetadata[msg.mergeRequestIid]
		if exists {
			metadata.mergeAutomatically = shouldBeMerged
			m.mrMetadata[msg.mergeRequestIid] = metadata
		}
		m.redrawTable()
	case MergeRequestProcessingResult:
		for mrIid, status := range msg.mrStatus {
			metadata, exists := m.mrMetadata[mrIid]
			if exists {
				metadata.status = status
				m.mrMetadata[mrIid] = metadata
			}
		}
		var toBeMerged = make(map[int]bool)
		for _, request := range m.mergeRequests {
			toBeMerged[request.Iid] = m.mrMetadata[request.Iid].mergeAutomatically == yes
		}
		cmds = append(cmds, m.processMergeRequests(toBeMerged))
		m.redrawTable()
	case context.UpdatedContextMessage:
		m.recalculateTable()
	}

	return m, tea.Batch(cmds...)
}

func (m *ActiveMergeRequestTable) redrawTable() {
	var rows []table.Row
	for _, mergeRequest := range m.mergeRequests {
		rows = append(rows, table.NewRow(table.RowData{
			columnKeyMergeRequest:         mergeRequest.Title,
			columnKeyMergeAutomatically:   m.mrMetadata[mergeRequest.Iid].mergeAutomatically,
			columnKeyStatus:               m.mrMetadata[mergeRequest.Iid].status,
			columnKeySourceBranch:         mergeRequest.SourceBranch,
			columnKeyTargetBranch:         mergeRequest.TargetBranch,
			columnKeyMergeRequestMetadata: mergeRequest,
		}))
	}
	m.flexTable = m.flexTable.WithRows(rows)
}

func (m *ActiveMergeRequestTable) recalculateTable() {
	m.flexTable = m.flexTable.WithTargetWidth(m.context.WindowWidth - m.context.Styles.Tabs.Content.GetHorizontalFrameSize())
	m.flexTable = m.flexTable.WithPageSize(m.context.TablePageSize)
}

func (m *ActiveMergeRequestTable) FullHelp() []key.Binding {
	return []key.Binding{}
}

func (m *ActiveMergeRequestTable) View() string {
	return m.flexTable.View()
}
