package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/romanyakovlev/gophkeeper/internal/client"
	"log"
	"strings"
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ElementData represents an element with ID, Name, and UserID
type ElementData struct {
	ID     string
	Name   string
	UserID string
	Type   string
}

// model represents the state of the TUI, including the elements and cursor positions
type model struct {
	elements              []ElementData
	cursorIndex           int
	actionCursor          int
	actionSelected        bool
	menuActionCursor      int
	menuActionSelected    bool
	menuActions           []string
	actions               []string
	logs                  []string
	credentialView        bool
	dataView              bool
	createCredentialsView bool
	loginCredentials      string
	passwordCredentials   string
	data                  string
}

var width, height, err = term.GetSize(0)

func initialModel() model {
	return model{
		elements:    []ElementData{},
		cursorIndex: 0,
		actions:     []string{"Get", "Delete"},
		menuActions: []string{"Create Credentials", "Exit"},
		logs:        []string{},
	}
}

// Msg for fetching elements
type fetchMsg struct {
	elements []ElementData
}

// Msg for clearing logs
type clearLogMsg int

func fetchElementsCmd() bubbletea.Cmd {
	return func() bubbletea.Msg {
		// Connect to your gRPC server
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer conn.Close()

		client := pb.NewKeeperServiceClient(conn)

		// Fetch the elements from the server
		res, err := client.GetElements(context.Background(), &emptypb.Empty{})
		if err != nil {
			return err
		}

		elements := make([]ElementData, len(res.Elements))
		for i, el := range res.Elements {
			elements[i] = ElementData{
				ID:     el.Id,
				Name:   el.Name,
				UserID: el.UserId,
				Type:   el.Type,
			}
		}

		return fetchMsg{elements: elements}
	}
}

func (m model) Init() bubbletea.Cmd {
	return fetchElementsCmd()
}

// custom message for logging
type logMsg string
type dataMsg string

// Encapsulate logging and creation of the delayed clear log command
func logToUI(msg string) bubbletea.Cmd {
	return func() bubbletea.Msg {
		return logMsg(msg)
	}
}

// Encapsulate logging and creation of the delayed clear log command
func dataToUI(msg string) bubbletea.Cmd {
	return func() bubbletea.Msg {
		return dataMsg(msg)
	}
}

// Encapsulate a delayed clear log command
func delayedClearLog(index int) bubbletea.Cmd {
	return func() bubbletea.Msg {
		time.Sleep(10 * time.Second)
		return clearLogMsg(index)
	}
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case fetchMsg:
		m.elements = msg.elements
		return m, nil
	case bubbletea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, bubbletea.Quit
		case "tab":
			m.menuActionSelected = true
		case "up", "k":
			if m.menuActionSelected {
				if m.menuActionCursor > 0 {
					m.menuActionCursor--
				}
			} else if m.actionSelected {
				if m.actionCursor > 0 {
					m.actionCursor--
				}
			} else {
				if m.cursorIndex > 0 {
					m.cursorIndex--
				}
			}
		case "down", "j":
			if m.menuActionSelected {
				if m.menuActionCursor < len(m.menuActions)-1 {
					m.menuActionCursor++
				}
			} else if m.actionSelected {
				if m.actionCursor < len(m.actions)-1 {
					m.actionCursor++
				}
			} else {
				if m.cursorIndex < len(m.elements)-1 {
					m.cursorIndex++
				}
			}
		case "enter", " ":
			if m.dataView {
				m.dataView = false
				//m.data = ""
			} else if m.credentialView {
				m.credentialView = false
				m.loginCredentials = ""
				m.passwordCredentials = ""
			} else if m.menuActionSelected {
				result := m.handleMenuAction()
				m.menuActionSelected = false
				m.menuActionCursor = 0
				return m, result
			} else if m.actionSelected {
				result := m.handleAction()
				m.actionSelected = false
				m.actionCursor = 0
				return m, result
			} else {
				m.actionSelected = true
			}
		}
	case logMsg:
		m.logs = append(m.logs, string(msg))
		return m, delayedClearLog(len(m.logs) - 1)
	case dataMsg:
		//m.dataView = true
		m.data = string(msg)
		return m, nil
	case clearLogMsg:
		index := int(msg)
		if index >= 0 && index < len(m.logs) {
			m.logs = append(m.logs[:index], m.logs[index+1:]...)
		}
	}

	return m, nil
}

func (m model) handleMenuAction() bubbletea.Cmd {
	switch m.menuActions[m.menuActionCursor] {
	case "Create Credentials":
		m.createCredentialsView = true
		return nil
	}
	return nil
}

func (m model) handleAction() bubbletea.Cmd {
	selectedElement := m.elements[m.cursorIndex]
	switch m.actions[m.actionCursor] {
	case "Get":
		if selectedElement.Type == "bytes" {
			conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()

			keeper := client.NewKeeperServiceClient(conn)

			filePath, err := keeper.GetBytes(context.Background(), selectedElement.ID)
			if err != nil {
				return logToUI(fmt.Sprintf("Failed to download: %v", err))
			}
			return dataToUI(fmt.Sprintf("Download successful, file saved as %s for object %s", filePath, selectedElement.ID))
		} else if selectedElement.Type == "credentials" {
			conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("Failed to connect: %v", err)
			}
			defer conn.Close()

			keeper := client.NewKeeperServiceClient(conn)

			resp, err := keeper.GetCredentials(context.Background(), selectedElement.ID)
			if err != nil {
				return logToUI(fmt.Sprintf("Failed to download: %v", err))
			}

			return dataToUI(
				fmt.Sprintf(
					"Credentials for object %s:\nLogin: %s\nPassword: %s",
					selectedElement.ID,
					resp.Login,
					resp.Password,
				),
			)
		}
	case "Delete":
		// Implement the delete functionality (currently empty)
		return logToUI(fmt.Sprintf("Delete action triggered for %s", selectedElement.ID))
	}
	return nil
}

func (m model) View() string {
	var sb strings.Builder
	/*
		if m.dataView {
			var sb strings.Builder
			sb.WriteString("Data:\n\n")
			sb.WriteString(m.data)
			return sb.String()
		}

	*/

	if m.createCredentialsView {
		style1 := lipgloss.NewStyle().Width(int(float64(width) * 0.6))
		style2 := lipgloss.NewStyle().Width(int(float64(width) * 0.4))

		view1 := m.viewMenuActionMenu()
		view2 := "Data:\n\n" + m.data

		return lipgloss.JoinHorizontal(lipgloss.Top, []string{style1.Render(view1), style2.Render(view2)}...)
	}

	if len(m.elements) == 0 {
		return "Loading..."
	}

	if m.menuActionSelected {
		style1 := lipgloss.NewStyle().Width(int(float64(width) * 0.6))
		style2 := lipgloss.NewStyle().Width(int(float64(width) * 0.4))

		view1 := m.viewMenuActionMenu()
		view2 := "Data:\n\n" + m.data

		return lipgloss.JoinHorizontal(lipgloss.Top, []string{style1.Render(view1), style2.Render(view2)}...)
	}

	if m.actionSelected {
		style1 := lipgloss.NewStyle().Width(int(float64(width) * 0.6))
		style2 := lipgloss.NewStyle().Width(int(float64(width) * 0.4))

		view1 := m.viewActionMenu()
		view2 := "Data:\n\n" + m.data

		return lipgloss.JoinHorizontal(lipgloss.Top, []string{style1.Render(view1), style2.Render(view2)}...)
	}

	// Define the table header
	headers := []string{"ID", "Name", "Type"}
	headerLine := fmt.Sprintf("%-36s  %-20s  %-20s", headers[0], headers[1], headers[2])

	// Helper function to format table rows
	formatRow := func(el ElementData, selected bool) string {
		row := fmt.Sprintf(
			"%-36s  %-20s  %-20s",
			el.ID, el.Name, el.Type,
		)
		if selected {
			return fmt.Sprintf("> %s <", row) // indicate selection
		}
		return row
	}

	sb.WriteString("Elements:\n\n")
	sb.WriteString(headerLine)
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", len(headerLine)))
	sb.WriteString("\n")

	for i, el := range m.elements {
		selected := i == m.cursorIndex
		sb.WriteString(formatRow(el, selected))
		sb.WriteString("\n")
	}
	sb.WriteString("\nPress Ctrl+C to exit.")
	sb.WriteString("\nPress Enter to select an item.")
	//sb.WriteString("\n\nData:\n\n")
	//sb.WriteString(m.data)
	/*
		sb.WriteString("\n\nLogs:\n")
		for _, logLine := range m.logs {
			sb.WriteString(logLine + "\n")
		}

	*/

	style1 := lipgloss.NewStyle().Width(int(float64(width) * 0.6))
	style2 := lipgloss.NewStyle().Width(int(float64(width) * 0.4))

	view1 := sb.String()
	view2 := "Data:\n\n" + m.data

	return lipgloss.JoinHorizontal(lipgloss.Top, []string{style1.Render(view1), style2.Render(view2)}...)
	//return
}

func (m model) createCredentialsViewMenu() string {
	var sb strings.Builder
	sb.WriteString("Select an action for the item:\n\n")

	for i, action := range m.actions {
		cursor := " "
		if m.actionCursor == i {
			cursor = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, action))
	}
	return sb.String()
}

func (m model) viewActionMenu() string {
	var sb strings.Builder
	sb.WriteString("Select an action for the item:\n\n")

	for i, action := range m.actions {
		cursor := " "
		if m.actionCursor == i {
			cursor = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, action))
	}
	return sb.String()
}

func (m model) viewMenuActionMenu() string {
	var sb strings.Builder
	sb.WriteString("Select an action for the item:\n\n")

	for i, action := range m.menuActions {
		cursor := " "
		if m.menuActionCursor == i {
			cursor = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, action))
	}
	return sb.String()
}

func main() {
	p := bubbletea.NewProgram(initialModel(), bubbletea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
