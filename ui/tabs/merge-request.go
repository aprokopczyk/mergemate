package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
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

type MergeRequestTable struct {
	flexTable     table.Model
	mrMetadata    map[int]RequestMetadata
	mergeRequests []gitlab.MergeRequestDetails
	totalMargin   int
	totalWidth    int
	gitlabClient  *gitlab.ApiClient
	keys          keys.MergeRequestKeyMap
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
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)).
			WithPageSize(10),
		gitlabClient: apiClient,
		totalMargin:  totalMargin,
		mrMetadata:   make(map[int]RequestMetadata),
		keys:         keys.MergeRequestHelp(),
	}
}

func (m *MergeRequestTable) listMergeRequests() tea.Msg {
	mergeRequests := m.gitlabClient.ListMergeRequests()
	return mergeRequests
}

func (m *MergeRequestTable) rebaseMergeRequest(mergeRequestIid int) tea.Cmd {
	return func() tea.Msg {
		pipelines, err := m.gitlabClient.GetMergeRequestPipelines(mergeRequestIid)
		if err != nil {
			log.Printf("Error when fetching piplelines for merge request {id = %v}: %v", mergeRequestIid, err)
			return nil
		}
		numberOfPipelines := len(pipelines)
		var shouldSkipCi = numberOfPipelines > 0 && pipelines[0].Status == "success"
		err = m.gitlabClient.RebaseMergeRequest(mergeRequestIid, shouldSkipCi)
		if err != nil {
			log.Printf("Error when rebasing merge request {id = %v}: %v", mergeRequestIid, err)
			return nil
		}
		return nil
	}
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

type AutomaticMergeResult struct {
	mrStatus map[int]string
}

const merged = "Merged"

func (m *MergeRequestTable) triggerAutomaticMerge(mergeRequestIids []int) tea.Cmd {
	return tea.Tick(time.Second*30, func(t time.Time) tea.Msg {
		log.Printf("Processing merge requests: %v", mergeRequestIids)
		mrStatus := make(map[int]string)
		var rebasing []int

		for _, mergeRequestIid := range mergeRequestIids {
			details, err := m.gitlabClient.GetMergeRequestDetails(mergeRequestIid)
			if err != nil {
				log.Printf("Fetching merge request details failed %v", err)
				continue
			}
			if details.RebaseInProgress {
				log.Printf("Merge request {id = %v, title=%v} is being rebased.", details.Iid, details.Title)
				mrStatus[mergeRequestIid] = "Rebase in progress"
				continue
			} else if details.RebaseError != "" {
				mrStatus[mergeRequestIid] = "Merge conflict"
				continue
			}
			pipelines, err := m.gitlabClient.GetMergeRequestPipelines(mergeRequestIid)
			if err != nil {
				log.Printf("Error when fetching pipeline for merge request{id = %v, title=%v}: %v", mergeRequestIid, details.Title, err)
			}
			if gitlab.IsPipelineRunning(pipelines) {
				mrStatus[mergeRequestIid] = "CI running"
				continue
			}
			if len(pipelines) > 0 && pipelines[0].Status == "failed" {
				mrStatus[mergeRequestIid] = "CI failed"
				continue
			}
			if details.CommitsBehind > 0 {
				log.Printf("Merge request {id = %v, title=%v} is behind target branch by %v commits, it will be rebased.", mergeRequestIid, details.Title, details.CommitsBehind)
				// we will rebase outside loop, in case there is mr that could be merged, we'll need to rebase only once
				rebasing = append(rebasing, mergeRequestIid)
				continue
			}
			if gitlab.IsAutomaticMergeAllowed(pipelines) {
				// hurray, we can merge it!
				log.Printf("Merging merge request {id = %v, title=%v}.", mergeRequestIid, details.Title)
				request, err := m.gitlabClient.MergeMergeRequest(mergeRequestIid)
				if err != nil {
					log.Printf("Error when merging merge request {id = %v, title=%v}: %v ", mergeRequestIid, details.Title, err)
					mrStatus[mergeRequestIid] = "Merge failed"
					return nil
				}
				if request.State == "merged" {
					log.Printf("Merged merge request {id = %v, title=%v}.", mergeRequestIid, details.Title)
					mrStatus[mergeRequestIid] = merged
				}
				continue
			}
		}

		for _, mrIid := range rebasing {
			mrStatus[mrIid] = "Rebase in progress"
			err := m.gitlabClient.RebaseMergeRequest(mrIid, true)
			if err != nil {
				log.Printf("Error when rebasing merge request {id = %v}: %v", mrIid, err)
			}
		}

		return AutomaticMergeResult{
			mrStatus: mrStatus,
		}
	})
}

func (m *MergeRequestTable) mergeMergeRequest(mergeRequestIid int) tea.Cmd {
	return func() tea.Msg {
		_, err := m.gitlabClient.MergeMergeRequest(mergeRequestIid)
		if err != nil {
			return nil
		}
		return nil
	}
}

func (m *MergeRequestTable) Init() tea.Cmd {
	return tea.Batch(m.listMergeRequests, m.triggerAutomaticMerge([]int{}))
}

func (m *MergeRequestTable) Update(msg tea.Msg) (TabContent, tea.Cmd) {
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
	case AutomaticMergeResult:
		for mrIid, status := range msg.mrStatus {
			metadata, exists := m.mrMetadata[mrIid]
			if exists {
				metadata.status = status
				m.mrMetadata[mrIid] = metadata
			}
		}
		var toBeMerged []int
		for _, request := range m.mergeRequests {
			if m.mrMetadata[request.Iid].mergeAutomatically == yes {
				toBeMerged = append(toBeMerged, request.Iid)
			}
		}
		cmds = append(cmds, m.triggerAutomaticMerge(toBeMerged))
		m.redrawTable()
	case tea.WindowSizeMsg:
		m.totalWidth = msg.Width
		m.recalculateTable()
		cmds = append(cmds, tea.ClearScreen)
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			mrToRebase := m.flexTable.HighlightedRow().Data[columnKeyMergeRequestMetadata].(gitlab.MergeRequestDetails)
			cmds = append(cmds, m.rebaseMergeRequest(mrToRebase.Iid))
		case "m":
			mrToMerge := m.flexTable.HighlightedRow().Data[columnKeyMergeRequestMetadata].(gitlab.MergeRequestDetails)
			cmds = append(cmds, m.mergeMergeRequest(mrToMerge.Iid))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *MergeRequestTable) redrawTable() {
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

func (m *MergeRequestTable) recalculateTable() {
	m.flexTable = m.flexTable.WithTargetWidth(m.totalWidth - m.totalMargin)
}

func (m *MergeRequestTable) FullHelp() []key.Binding {
	return []key.Binding{
		m.keys.Rebase,
		m.keys.Merge,
	}
}

func (m *MergeRequestTable) View() string {
	return m.flexTable.View() + "\n"
}
