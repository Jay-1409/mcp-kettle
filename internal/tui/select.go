package tui

import (
	"fmt"
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
	cancelled  bool
}

func Select(candidates []model.Candidate) ([]model.Candidate, bool, error) {
	selected := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		selected[candidate.ID] = true
	}
	m := selectionModel{candidates: candidates, selected: selected}
	m.list = list.New(m.items(), list.NewDefaultDelegate(), 100, 24)
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
			if current, ok := m.list.SelectedItem().(item); ok {
				m.selected[current.candidate.ID] = !m.selected[current.candidate.ID]
				return m, m.list.SetItems(m.items())
			}
		case "a":
			m.setAll(true)
			return m, m.list.SetItems(m.items())
		case "n":
			m.setAll(false)
			return m, m.list.SetItems(m.items())
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectionModel) View() tea.View {
	count := m.selectedCount()
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		fmt.Sprintf("  %d/%d selected · space toggle · a all · n none · / filter · enter generate · esc cancel", count, len(m.candidates)),
	)
	view := tea.NewView(m.list.View() + "\n" + footer)
	view.AltScreen = true
	return view
}

func (m selectionModel) items() []list.Item {
	items := make([]list.Item, 0, len(m.candidates))
	for _, candidate := range m.candidates {
		items = append(items, item{candidate: candidate, selected: m.selected[candidate.ID]})
	}
	return items
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
