package tabs

import (
	"github.com/aprokopczyk/mergemate/pkg/gitlab"
	"github.com/aprokopczyk/mergemate/ui/colors"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"sort"
	"strings"
)

const (
	columnKeyBranchName     = "branchName"
	columnKeyLastCommit     = "lastCommit"
	columnKeyBranchMetadata = "branchDetails"
)

type BranchTable struct {
	branchesList     list.Model
	flexTable        table.Model
	keys             keyMap
	help             help.Model
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

type keyMap struct {
	Up                      key.Binding
	Down                    key.Binding
	Left                    key.Binding
	Right                   key.Binding
	MergeAutomatically      key.Binding
	CloseTargetBranchesList key.Binding
	SelectTargetBranch      key.Binding
}

var keys = keyMap{
	Up:                      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:                    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Left:                    key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "Switch to left tab")),
	Right:                   key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "Switch to right tab")),
	MergeAutomatically:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "Create automatic merge request")),
	CloseTargetBranchesList: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Close target branches list"), key.WithDisabled()),
	SelectTargetBranch:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Select target branch"), key.WithDisabled()),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Down}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.MergeAutomatically, k.CloseTargetBranchesList, k.SelectTargetBranch},
	}
}

func NewBranchTable(apiClient *gitlab.ApiClient, totalMargin int) *BranchTable {
	helpModel := help.New()
	helpModel.ShowAll = true
	return &BranchTable{
		flexTable: table.New([]table.Column{
			table.NewFlexColumn(columnKeyBranchName, "Branch", 15),
			table.NewFlexColumn(columnKeyLastCommit, "Last commit date", 15),
		}).WithRows([]table.Row{}).
			Focused(true).
			HeaderStyle(lipgloss.NewStyle().Bold(true)).
			WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left).BorderForeground(colors.Emerald600)).
			WithPageSize(10),
		branchesList:     createList(),
		keys:             keys,
		help:             helpModel,
		gitlabClient:     apiClient,
		totalMargin:      totalMargin,
		branches:         []gitlab.Branch{},
		showMergeTargets: false,
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

	sort.SliceStable(branches, func(i, j int) bool {
		return branches[i].Commit.AuthoredDate.Unix() > branches[j].Commit.AuthoredDate.Unix()
	})

	return branches
}

func (m *BranchTable) createMergeRequest(sourceBranch string, targetBranch string, title string) tea.Cmd {
	return func() tea.Msg {
		mrIid := m.gitlabClient.CreateMergeRequest(sourceBranch, targetBranch, title)
		m.gitlabClient.CreateMergeRequestNote(mrIid, MergeAutomatically)
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

	switch msg := msg.(type) {
	case []gitlab.Branch:
		var rows []table.Row
		var targetBranches []list.Item
		for i := 0; i < len(msg); i++ {
			if strings.HasPrefix(msg[i].Name, m.gitlabClient.BranchPrefix) {
				rows = append(rows, table.NewRow(table.RowData{
					columnKeyBranchName:     msg[i].Name,
					columnKeyLastCommit:     msg[i].Commit.AuthoredDate,
					columnKeyBranchMetadata: msg[i],
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
		switch {
		case key.Matches(msg, m.keys.MergeAutomatically):
			if !m.showMergeTargets {
				m.changeBranchSelectionVisibility(true)
			}
		case key.Matches(msg, m.keys.CloseTargetBranchesList):
			if m.showMergeTargets && m.branchesList.FilterState() != list.Filtering {
				m.changeBranchSelectionVisibility(false)
			}
		case key.Matches(msg, m.keys.SelectTargetBranch):
			if m.showMergeTargets && m.branchesList.FilterState() != list.Filtering {
				sourceBranch := m.flexTable.HighlightedRow().Data[columnKeyBranchMetadata].(gitlab.Branch)
				targetBranch := m.branchesList.SelectedItem().(branchItem)
				cmds = append(cmds, m.createMergeRequest(sourceBranch.Name, targetBranch.name, sourceBranch.Commit.Message))
				m.changeBranchSelectionVisibility(false)
			}
		}
	}

	if !m.showMergeTargets {
		m.flexTable, cmd = m.flexTable.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.branchesList, cmd = m.branchesList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *BranchTable) changeBranchSelectionVisibility(visible bool) {
	m.keys.CloseTargetBranchesList.SetEnabled(visible)
	m.keys.SelectTargetBranch.SetEnabled(visible)
	m.keys.MergeAutomatically.SetEnabled(!visible)
	m.showMergeTargets = visible
	m.recalculateComponents()
}

func (m *BranchTable) recalculateComponents() {
	tableWidth := m.tableSize()
	m.flexTable = m.flexTable.WithTargetWidth(tableWidth)
	m.help.Width = tableWidth
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
		return lipgloss.JoinHorizontal(lipgloss.Top, m.flexTable.View()+"\n"+m.help.View(m.keys), m.branchesList.View())
	}
	return m.flexTable.View() + "\n" + m.help.View(m.keys)
}
