package main

// An example demonstrating an application with multiple views.
//
// Note that this example was produced before the Bubbles progress component
// was available (github.com/charmbracelet/bubbles/progress) and thus, we're
// implementing a progress bar from scratch here.

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	progressBarWidth  = 71
	progressFullChar  = "█"
	progressEmptyChar = "░"
	dotChar           = " • "
)

// General stuff for styling the view
var (
	keywordStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ticksStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("79"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	progressEmpty = subtleStyle.Render(progressEmptyChar)
	dotStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)

	// Gradient colors we'll use for the progress bar
	ramp = makeRampStyles("#B14FFF", "#00FFA3", progressBarWidth)
)

var menuActions = map[int]string{
	0: "Create Credit Card",
	1: "Create Credentials",
	2: "Create Bytes data from file",
	3: "Exit",
}

type (
	tickMsg  struct{}
	frameMsg struct{}
)

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

type mainMenuModel struct {
	Choice      int
	Ticks       int
	Frames      int
	Progress    float64
	Loaded      bool
	Quitting    bool
	parentModel *model
}

func initialCreateMainMenuModel(parentModel model) mainMenuModel {
	return mainMenuModel{0, 10, 0, 0, false, false, &parentModel}
}

func (m mainMenuModel) Init() tea.Cmd {
	return tick()
}

// Main update function.
func (m mainMenuModel) Update(msg tea.Msg) (mainMenuModel, tea.Cmd) {
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
	return updateChoices(msg, m)
}

// The main view, which just calls the appropriate sub-view
func (m mainMenuModel) View() string {
	var s string
	if m.Quitting {
		return "\n  See you later!\n\n"
	}
	s = choicesView(m)
	return mainStyle.Render(s)
}

// Sub-update functions

func (m mainMenuModel) handleMenuAction() tea.Cmd {
	action, _ := menuActions[m.Choice]
	switch action {
	case "Create Credentials":
		return createCredentials()
	case "Create Credit Card":
		return createCreditCard()
	case "Create Bytes data from file":
		return createBytes()
	}
	return nil
}

// Update loop for the first view where you're choosing a task.
func updateChoices(msg tea.Msg, m mainMenuModel) (mainMenuModel, tea.Cmd) {
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

	case tickMsg:
		if m.Ticks == 0 {
			m.Quitting = true
			return m, tea.Quit
		}
		m.Ticks--
		return m, tick()
	}

	return m, nil
}

// Sub-views

// The first view, where you're choosing a task
func choicesView(m mainMenuModel) string {
	c := m.Choice

	tpl := "Menu\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render("j/k, up/down: select") + dotStyle +
		subtleStyle.Render("enter: choose") + dotStyle +
		subtleStyle.Render("q, esc: quit")
	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		checkbox("Create Credit Card", c == 0),
		checkbox("Create Credentials", c == 1),
		checkbox("Create Bytes data from file", c == 2),
		checkbox("Exit", c == 3),
	)
	return fmt.Sprintf(tpl, choices)
}
