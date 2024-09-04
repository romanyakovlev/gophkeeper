package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/romanyakovlev/gophkeeper/internal/client"
	"google.golang.org/grpc"
	"log"
	"strings"
)

type CreateCredentialsModel struct {
	focusIndex  int
	inputs      []textinput.Model
	cursorMode  cursor.Mode
	parentModel *model
}

type credentialsView bool

func createCredentials() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return credentialsView(true)
	}
}

type credentialsCreatedType bool

func credentialsCreated() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return credentialsCreatedType(true)
	}
}

func initialCreateCredentialsModel(parentModel model) CreateCredentialsModel {
	m := CreateCredentialsModel{
		inputs:      make([]textinput.Model, 2),
		parentModel: &parentModel,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Nickname"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m CreateCredentialsModel) Init() bubbletea.Cmd {
	return textinput.Blink
}

func (m CreateCredentialsModel) handleAction() bubbletea.Cmd {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	keeper := client.NewKeeperServiceClient(conn)

	_, err = keeper.SaveCredentials(context.Background(), m.inputs[0].Value(), m.inputs[1].Value())
	if err != nil {
		return logToUI(fmt.Sprintf("Failed to create: %v", err))
	}

	return credentialsCreated()
}

func (m CreateCredentialsModel) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case credentialsCreatedType:
		return m.parentModel, credentialsCreated()
	case logMsg:
		return m.parentModel, logToUI(string(msg))
	case bubbletea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, bubbletea.Quit
		case "esc":
			return m.parentModel, m.parentModel.Init()

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]bubbletea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, bubbletea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, m.handleAction()
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]bubbletea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, bubbletea.Batch(cmds...)

		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *CreateCredentialsModel) updateInputs(msg bubbletea.Msg) bubbletea.Cmd {
	cmds := make([]bubbletea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return bubbletea.Batch(cmds...)
}

func (m CreateCredentialsModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
