package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type selectionItem[T comparable] struct {
	label string
	value T
}

type selectionModel[T comparable] struct {
	title    string
	items    []selectionItem[T]
	cursor   int
	chose    bool
	selected T
}

func (m selectionModel[T]) Init() tea.Cmd { return nil }

func (m selectionModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, len(m.items)-1)
	case "enter":
		if len(m.items) == 0 {
			return m, tea.Quit
		}
		m.selected = m.items[m.cursor].value
		m.chose = true
		return m, tea.Quit
	}
	return m, nil
}

func (m selectionModel[T]) View() string {
	var out strings.Builder
	if m.title != "" {
		out.WriteString(m.title + "\n\n")
	}

	for i, item := range m.items {
		if i == m.cursor {
			out.WriteString(focusStyle.Render("> "+item.label) + "\n")
			continue
		}
		out.WriteString("  " + item.label + "\n")
	}

	out.WriteString("\n" + mutedStyle.Render("(↑/↓ move, enter select, esc cancel)") + "\n")
	return out.String()
}

func runSelection[T comparable](title string, items []selectionItem[T]) (T, bool, error) {
	final, err := tea.NewProgram(selectionModel[T]{title: title, items: items}).Run()
	if err != nil {
		var zero T
		return zero, false, err
	}

	result := final.(selectionModel[T])
	if !result.chose {
		var zero T
		return zero, false, nil
	}
	return result.selected, true, nil
}

type textPromptModel struct {
	title      string
	input      textinput.Model
	validate   func(string) error
	confirmed  bool
	errMessage string
}

func newTextPromptModel(title, prompt, initial string, validate func(string) error) textPromptModel {
	input := textinput.New()
	input.Prompt = prompt
	input.SetValue(initial)
	input.Focus()
	return textPromptModel{title: title, input: input, validate: validate}
}

func (m textPromptModel) Init() tea.Cmd { return textinput.Blink }

func (m textPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.validate != nil {
				if err := m.validate(strings.TrimSpace(m.input.Value())); err != nil {
					m.errMessage = err.Error()
					return m, nil
				}
			}
			m.confirmed = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.errMessage = ""
	return m, cmd
}

func (m textPromptModel) View() string {
	var out strings.Builder
	if m.title != "" {
		out.WriteString(m.title + "\n\n")
	}
	out.WriteString(m.input.View() + "\n")
	if m.errMessage != "" {
		out.WriteString(warnStyle.Render(m.errMessage) + "\n")
	}
	out.WriteString(mutedStyle.Render("(enter to continue, esc to cancel)") + "\n")
	return out.String()
}

func runTextPrompt(title, prompt, initial string, validate func(string) error) (string, bool, error) {
	final, err := tea.NewProgram(newTextPromptModel(title, prompt, initial, validate)).Run()
	if err != nil {
		return "", false, err
	}

	result := final.(textPromptModel)
	if !result.confirmed {
		return "", false, nil
	}
	return strings.TrimSpace(result.input.Value()), true, nil
}
