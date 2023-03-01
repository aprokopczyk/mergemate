package styles

import (
	"github.com/aprokopczyk/mergemate/ui/colors"
	"github.com/charmbracelet/lipgloss"
)

const (
	TabsHeaderHeight  = 3
	TableHeaderHeight = 3
	TableFooterHeight = 3
	MinTablePageSize  = 5
)

type Styles struct {
	Tabs struct {
		TabItem lipgloss.Style
		Header  lipgloss.Style
		Content lipgloss.Style
	}
	Help lipgloss.Style
}

func NewStyles() Styles {
	var styles Styles

	styles.Tabs.TabItem = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		Padding(0, 1, 0, 1).
		BorderForeground(colors.Emerald800)
	styles.Tabs.Header = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), false, false, true, false).
		Padding(1, 0, 0, 2).
		BorderForeground(colors.Emerald800)
	styles.Tabs.Content = lipgloss.NewStyle().Padding(0, 0, 0, 2)

	styles.Help = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true, false, false, false).
		BorderForeground(colors.Emerald800)

	return styles
}
