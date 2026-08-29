package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"mcp-kettel/internal/model"
)

type item struct {
	candidate model.Candidate
	selected  bool
}

type groupItem struct {
	name     string
	ids      []string
	expanded bool
	focused  bool
}

func (i groupItem) Title() string {
	marker := "▸"
	if i.expanded {
		marker = "▾"
	}
	return fmt.Sprintf("%s %s (%d)", marker, i.name, len(i.ids))
}
func (i groupItem) Description() string {
	if i.focused {
		return "space expands or collapses this group"
	}
	return ""
}
func (i groupItem) FilterValue() string { return i.name }

type grouping uint8

const (
	groupNone grouping = iota
	groupSource
	groupMethod
	groupPath
)

func (g grouping) String() string {
	switch g {
	case groupSource:
		return "source"
	case groupMethod:
		return "method"
	case groupPath:
		return "path"
	default:
		return "none"
	}
}

func (i item) Title() string {
	mark := "○"
	if i.selected {
		mark = "●"
	}
	return fmt.Sprintf("%s  %s", mark, i.candidate.Label())
}

func (i item) Description() string {
	return i.candidate.ToolName + " · " + i.candidate.Description
}

func (i item) FilterValue() string {
	return strings.Join([]string{i.candidate.ToolName, i.candidate.Method, i.candidate.Route, i.candidate.SourceFile}, " ")
}

type selectionModel struct {
	list       list.Model
	candidates []model.Candidate
	selected   map[string]bool
	grouping   grouping
	expanded   map[string]bool
	focused    string
	cancelled  bool
}

func Select(candidates []model.Candidate) ([]model.Candidate, bool, error) {
	selected := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		selected[candidate.ID] = true
	}
	m := selectionModel{candidates: candidates, selected: selected, expanded: make(map[string]bool)}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.FilterMatch = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F793FF")).
		Background(lipgloss.NoColor{})
	m.list = list.New(m.items(), delegate, 100, 24)
	m.list.Title = "MCP Kettel · choose API tools"
	m.list.AdditionalFullHelpKeys = nil

	result, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, false, err
	}
	final := result.(selectionModel)
	if final.cancelled {
		return nil, true, nil
	}
	var chosen []model.Candidate
	for _, candidate := range candidates {
		if final.selected[candidate.ID] {
			chosen = append(chosen, candidate)
		}
	}
	return chosen, false, nil
}

func (m selectionModel) Init() tea.Cmd { return nil }

func (m selectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.list.SetSize(size.Width, size.Height-2)
	}
	if key, ok := msg.(tea.KeyPressMsg); ok && !m.list.SettingFilter() {
		switch key.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if m.selectedCount() > 0 {
				return m, tea.Quit
			}
		case "space":
			switch current := m.list.SelectedItem().(type) {
			case item:
				m.selected[current.candidate.ID] = !m.selected[current.candidate.ID]
			case groupItem:
				m.expanded[current.name] = !current.expanded
			default:
				return m, nil
			}
			return m, m.list.SetItems(m.items())
		case "a":
			m.setAll(true)
			return m, m.list.SetItems(m.items())
		case "n":
			m.setAll(false)
			return m, m.list.SetItems(m.items())
		case "g":
			m.grouping = (m.grouping + 1) % 4
			m.expanded = make(map[string]bool)
			m.focused = ""
			return m, m.list.SetItems(m.items())
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if current, ok := m.list.SelectedItem().(groupItem); ok {
		m.focused = current.name
	} else {
		m.focused = ""
	}
	if m.grouping != groupNone {
		cmd = tea.Sequence(cmd, m.list.SetItems(m.items()))
	}
	return m, cmd
}

func (m selectionModel) View() tea.View {
	count := m.selectedCount()
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		fmt.Sprintf("  %d/%d selected · g group: %s · %s · a all · n none · / filter · enter generate · esc cancel", count, len(m.candidates), m.grouping, m.spaceHint()),
	)
	view := tea.NewView(m.list.View() + "\n" + footer)
	view.AltScreen = true
	return view
}

func (m selectionModel) spaceHint() string {
	if _, ok := m.list.SelectedItem().(groupItem); ok {
		return "space expand/collapse"
	}
	return "space toggle"
}

func (m selectionModel) items() []list.Item {
	if m.grouping == groupNone {
		return m.candidateItems(m.candidates)
	}
	grouped := append([]model.Candidate(nil), m.candidates...)
	sort.SliceStable(grouped, func(i, j int) bool {
		ki, kj := m.groupKey(grouped[i]), m.groupKey(grouped[j])
		if ki == kj {
			return grouped[i].ID < grouped[j].ID
		}
		return ki < kj
	})
	items := make([]list.Item, 0, len(grouped)+len(grouped))
	for index := 0; index < len(grouped); {
		name := m.groupKey(grouped[index])
		group := grouped[index : index+1]
		for len(group) < len(grouped)-index && m.groupKey(grouped[index+len(group)]) == name {
			group = grouped[index : index+len(group)+1]
		}
		ids := make([]string, len(group))
		for i, candidate := range group {
			ids[i] = candidate.ID
		}
		expanded := m.expanded[name]
		items = append(items, groupItem{name: name, ids: ids, expanded: expanded, focused: name == m.focused})
		if expanded {
			items = append(items, m.candidateItems(group)...)
		}
		index += len(group)
	}
	return items
}

func (m selectionModel) candidateItems(candidates []model.Candidate) []list.Item {
	items := make([]list.Item, 0, len(m.candidates))
	for _, candidate := range candidates {
		items = append(items, item{candidate: candidate, selected: m.selected[candidate.ID]})
	}
	return items
}

func (m selectionModel) groupKey(candidate model.Candidate) string {
	switch m.grouping {
	case groupSource:
		return candidate.SourceFile
	case groupMethod:
		return candidate.Method
	case groupPath:
		path := strings.Trim(candidate.Route, "/")
		if path == "" {
			return "/"
		}
		if index := strings.IndexByte(path, '/'); index >= 0 {
			path = path[:index]
		}
		return "/" + path + "/*"
	default:
		return ""
	}
}

func (m selectionModel) setAll(selected bool) {
	for _, candidate := range m.candidates {
		m.selected[candidate.ID] = selected
	}
}

func (m selectionModel) selectedCount() int {
	count := 0
	for _, selected := range m.selected {
		if selected {
			count++
		}
	}
	return count
}
