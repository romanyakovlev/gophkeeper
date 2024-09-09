package main

// An example demonstrating an application with multiple views.
//
// Note that this example was produced before the Bubbles progress component
// was available (github.com/charmbracelet/bubbles/progress) and thus, we're
// implementing a progress bar from scratch here.

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
)

var actionMenuActions = map[int]string{
	0: "Get",
	1: "Delete",
}

type actionMenuModel struct {
	Choice      int
	Ticks       int
	Frames      int
	Progress    float64
	Loaded      bool
	Quitting    bool
	parentModel *model
}

func initialCreateActionMenuModel(parentModel model) actionMenuModel {
	return actionMenuModel{0, 10, 0, 0, false, false, &parentModel}
}

func (m actionMenuModel) Init() tea.Cmd {
	return tick()
}

// Main update function.
func (m actionMenuModel) Update(msg tea.Msg) (actionMenuModel, tea.Cmd) {
	// Make sure these keys always quit

	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "esc" {
			m.parentModel.menuActionSelected = false
			return m, exitFromMenuCmd()
		}
		if k == "q" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	return m.updateChoices(msg)
}

// The main view, which just calls the appropriate sub-view
func (m actionMenuModel) View() string {
	var s string
	if m.Quitting {
		return "\n  See you later!\n\n"
	}
	s = m.choicesView()
	return mainStyle.Render(s)
}

// Sub-update functions

func (m actionMenuModel) handleMenuAction() tea.Cmd {
	action, _ := actionMenuActions[m.Choice]
	switch action {
	case "Get":
		return getElementDataCmd()
	case "Delete":
		return deleteElementDataCmd()
	}
	return nil
}

// Update loop for the first view where you're choosing a task.
func (m actionMenuModel) updateChoices(msg tea.Msg) (actionMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.Choice++
			if m.Choice > 3 {
				m.Choice = 3
			}
		case "k", "up":
			m.Choice--
			if m.Choice < 0 {
				m.Choice = 0
			}
		case "enter":
			return m, m.handleMenuAction()
		}
	}
	return m, nil
}

// Sub-views

// The first view, where you're choosing a task
func (m actionMenuModel) choicesView() string {
	c := m.Choice

	tpl := "Menu\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("j/k, up/down: select") + dotStyle +
		subtleStyle.Render("enter: choose") + dotStyle +
		subtleStyle.Render("q, esc: quit")
	choices := fmt.Sprintf(
		"%s\n%s",
		checkbox("Get", c == 0),
		checkbox("Delete", c == 1),
	)
	action, _ := actionMenuActions[m.Choice]
	return fmt.Sprintf(tpl, choices) + "\n\n" + strconv.FormatInt(int64(m.Choice), 10) + "  " + action
}
