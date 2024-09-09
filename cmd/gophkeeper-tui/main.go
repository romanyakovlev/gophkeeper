package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/romanyakovlev/gophkeeper/internal/client"
	pb "github.com/romanyakovlev/gophkeeper/internal/protobuf/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"strings"
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"
)

// model represents the state of the TUI, including the elements and cursor positions
type model struct {
	elements           list.Model
	cursorIndex        int
	actionCursor       int
	actionSelected     bool
	menuActionCursor   int
	menuActionSelected bool
	menuActions        []string
	logs               []string
	data               string

	// credentials

	createCredentialsView     bool
	credentialsModel          CreateCredentialsModel
	createCredentialsViewBool bool

	// credit card info

	createCreditCardView     bool
	creditCardModel          CreateCreditCardModel
	createCreditCardViewBool bool

	// credit card info

	createBytesCardView bool
	bytesModel          CreateBytesModel
	createBytesViewBool bool

	choice   string
	quitting bool

	// Menu model

	mainMenu   mainMenuModel
	actionMenu actionMenuModel
}

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

var width, height, err = term.GetSize(0)

func initialModel() model {
	const defaultWidth = 20

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), defaultWidth, listHeight)

	m := model{
		elements:    l,
		cursorIndex: 0,
		logs:        []string{},
	}
	m.mainMenu = initialCreateMainMenuModel(m)
	m.actionMenu = initialCreateActionMenuModel(m)
	m.credentialsModel = initialCreateCredentialsModel(m)
	m.creditCardModel = initialCreateCreditCardModel(m)
	m.bytesModel = initialCreateBytesModel(m)
	return m
}

// Msg for fetching elements
type fetchMsg struct {
	//elements []ElementData
	elements list.Model
}

type getElementData bool

func getElementDataCmd() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return getElementData(true)
	}
}

type deleteElementData bool

func deleteElementDataCmd() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return deleteElementData(true)
	}
}

// Msg for clearing logs
type clearLogMsg int

type exitFromMenu bool

func exitFromMenuCmd() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return exitFromMenu(true)
	}
}

func fetchElementsCmd() bubbletea.Cmd {
	return func() bubbletea.Msg {

		// Connect to your gRPC server
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer conn.Close()

		client := pb.NewKeeperServiceClient(conn)

		res, err := client.GetElements(context.Background(), &emptypb.Empty{})
		if err != nil {
			return err
		}

		elements := make([]list.Item, len(res.Elements))
		for i, el := range res.Elements {
			elements[i] = item{
				ID:     el.Id,
				Name:   el.Name,
				UserID: el.UserId,
				Type:   el.Type,
			}
		}

		const defaultWidth = 100
		l := list.New(elements, list.NewDefaultDelegate(), defaultWidth, listHeight)
		l.Title = "Gophkeeper items"

		return fetchMsg{elements: l}

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
	case exitFromMenu:
		m.mainMenu = initialCreateMainMenuModel(m)
		m.actionMenu = initialCreateActionMenuModel(m)
		m.menuActionSelected = false
		m.actionSelected = false
		return m, m.Init()
	case fetchMsg:
		m.elements = msg.elements
		return m, nil
	case getElementData:
		m.menuActionSelected = false
		m.actionSelected = false
		return m, m.handleGetAction()
	case deleteElementData:
		m.menuActionSelected = false
		m.actionSelected = false
		return m, m.handleDeleteAction()
	// credentials
	case credentialsView:
		m.credentialsModel = initialCreateCredentialsModel(m)
		m.createCredentialsViewBool = true
		m.menuActionSelected = false
		m.actionSelected = false
	case credentialsCreatedType:
		m.credentialsModel = initialCreateCredentialsModel(m)
		m.mainMenu = initialCreateMainMenuModel(m)
		m.actionMenu = initialCreateActionMenuModel(m)
		m.createCredentialsViewBool = false
		m.menuActionSelected = false
		m.actionSelected = false
		m.data = "Credentials object is created"
		//return m, m.Init()
	// credit card
	case creditCardView:
		m.creditCardModel = initialCreateCreditCardModel(m)
		m.createCreditCardViewBool = true
		m.menuActionSelected = false
	case creditCardCreatedType:
		m.creditCardModel = initialCreateCreditCardModel(m)
		m.mainMenu = initialCreateMainMenuModel(m)
		m.actionMenu = initialCreateActionMenuModel(m)
		m.createCreditCardViewBool = false
		m.menuActionSelected = false
		m.actionSelected = false
		m.data = "Credit Card object is created"
		return m, m.Init()
	// bytes
	case bytesView:
		m.bytesModel = initialCreateBytesModel(m)
		m.createBytesViewBool = true
		m.menuActionSelected = false
		return m, m.bytesModel.Init()
	case bytesCreatedType:
		m.bytesModel = initialCreateBytesModel(m)
		m.mainMenu = initialCreateMainMenuModel(m)
		m.actionMenu = initialCreateActionMenuModel(m)
		m.createBytesViewBool = false
		m.menuActionSelected = false
		m.actionSelected = false
		m.data = "Bytes(text) data object as file is created"
		return m, m.Init()
	case bubbletea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.elements.SetSize(msg.Width-h, msg.Height-v)
	case bubbletea.KeyMsg:
		if m.createCredentialsViewBool {
			return m.credentialsModel.Update(msg)
		}
		if m.createCreditCardViewBool {
			return m.creditCardModel.Update(msg)
		}
		var cmd bubbletea.Cmd
		if m.createBytesViewBool {
			m.bytesModel, cmd = m.bytesModel.Update(msg)
			return m, cmd
		}
		if m.menuActionSelected {
			m.mainMenu, cmd = m.mainMenu.Update(msg)
			return m, cmd
		} else if m.actionSelected {
			m.actionMenu, cmd = m.actionMenu.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, bubbletea.Quit
		case "tab":
			m.menuActionSelected = true
			return m, nil
		case "enter", " ":
			m.actionSelected = true
			return m, nil
		}
	case logMsg:
		m.logs = append(m.logs, string(msg))
		return m, delayedClearLog(len(m.logs) - 1)
	case dataMsg:
		m.data = string(msg)
		return m, m.Init()
	case clearLogMsg:
		index := int(msg)
		if index >= 0 && index < len(m.logs) {
			m.logs = append(m.logs[:index], m.logs[index+1:]...)
		}
	}
	var cmd bubbletea.Cmd
	if m.createBytesViewBool {
		m.bytesModel, cmd = m.bytesModel.Update(msg)
		return m, cmd
	}
	if m.menuActionSelected {
		m.mainMenu, cmd = m.mainMenu.Update(msg)
		return m, cmd
	} else if m.actionSelected {
		m.actionMenu, cmd = m.actionMenu.Update(msg)
		return m, cmd
	} else {
		m.elements, cmd = m.elements.Update(msg)
		return m, cmd
	}
}

func (m model) handleGetAction() bubbletea.Cmd {
	selectedElement, _ := m.elements.SelectedItem().(item)
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
	} else if selectedElement.Type == "card" {
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		keeper := client.NewKeeperServiceClient(conn)

		resp, err := keeper.GetCreditCard(context.Background(), selectedElement.ID)
		if err != nil {
			return logToUI(fmt.Sprintf("Failed to download: %v", err))
		}

		return dataToUI(
			fmt.Sprintf(
				"Credit Card, object %s:\nNumber: %s\nEXP: %s\nEXP: %s",
				selectedElement.ID,
				resp.CardNumber,
				resp.Exp,
				resp.Cvv,
			),
		)
	}
	return nil
}

func (m model) handleDeleteAction() bubbletea.Cmd {
	selectedElement, _ := m.elements.SelectedItem().(item)
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	keeper := client.NewKeeperServiceClient(conn)
	if selectedElement.Type == "bytes" {
		_, err := keeper.DeleteBytes(context.Background(), selectedElement.ID)
		if err != nil {
			return logToUI(fmt.Sprintf("Failed to delete bytes: %v", err))
		}
		return dataToUI(fmt.Sprintf("Delete for object %s", selectedElement.ID))
	} else if selectedElement.Type == "credentials" {
		_, err := keeper.DeleteCredentials(context.Background(), selectedElement.ID)
		if err != nil {
			return logToUI(fmt.Sprintf("Failed to delete credentials: %v", err))
		}
		return dataToUI(fmt.Sprintf("Delete for object %s", selectedElement.ID))
	} else if selectedElement.Type == "card" {
		_, err := keeper.DeleteCreditCard(context.Background(), selectedElement.ID)
		if err != nil {
			return logToUI(fmt.Sprintf("Failed to delete credit card: %v", err))
		}
		return dataToUI(fmt.Sprintf("Delete for object %s", selectedElement.ID))
	}

	return nil
}

func (m model) View() string {
	var sb strings.Builder

	sb.WriteString("\nPress Tab for menu.")
	sb.WriteString("\nPress Ctrl+C to exit.")
	sb.WriteString("\nPress Enter to select an item.")

	var view1 string
	if m.actionSelected {
		view1 = m.actionMenu.View() + docStyle.MarginLeft(2).Render(sb.String())
	} else if m.menuActionSelected {
		view1 = m.mainMenu.View() + docStyle.MarginLeft(2).Render(sb.String())
	} else if m.createCredentialsViewBool {
		return m.credentialsModel.View()
	} else if m.createCreditCardViewBool {
		return m.creditCardModel.View()
	} else if m.createBytesViewBool {
		view1 = m.bytesModel.View() + docStyle.MarginLeft(2).Render(sb.String())
	} else {
		view1 = m.elements.View() + docStyle.MarginLeft(2).Render(sb.String())
	}
	view2 := m.elements.Styles.Title.Render("Data") + "\n\n" + m.data

	style1 := docStyle.Width(int(float64(width) * 0.5))
	style2 := lipgloss.NewStyle().Margin(2, 2, 2, 0).Width(int(float64(width) * 0.5))

	return lipgloss.JoinHorizontal(lipgloss.Top, []string{style1.Render(view1), style2.Render(view2)}...)
}

func main() {
	//p := bubbletea.NewProgram(initialModel(), bubbletea.WithAltScreen()) // caught some bugs with list.Model rendering
	p := bubbletea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
