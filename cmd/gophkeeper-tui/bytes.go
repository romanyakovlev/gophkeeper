package main

import (
	"context"
	"errors"
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/romanyakovlev/gophkeeper/internal/client"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
	"time"
)

type CreateBytesModel struct {
	filepicker   filepicker.Model
	parentModel  *model
	selectedFile string
	quitting     bool
	err          error
	initialized  bool
	uploading    bool
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m CreateBytesModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m *CreateBytesModel) handleAction() tea.Cmd {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	keeper := client.NewKeeperServiceClient(conn)
	err = keeper.SaveBytes(context.Background(), m.selectedFile)
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}
	return bytesCreated()
}

func (m *CreateBytesModel) Update(msg tea.Msg) (CreateBytesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return *m, tea.Quit
		}
	case clearErrorMsg:
		m.err = nil
	case uploading:
		m.uploading = false
		return *m, m.handleAction()
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
		m.uploading = true
		return *m, createUploading()
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return *m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return *m, cmd
}

func (m CreateBytesModel) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("Select file to upload" + "\n\n")
	if m.err != nil {
		s.WriteString("error\n\n")
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else if m.uploading {
		return lipgloss.NewStyle().Margin(2).Render("Uploading, please wait...")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}

type bytesView bool

func createBytes() tea.Cmd {
	return func() tea.Msg {
		return bytesView(true)
	}
}

type uploading bool

func createUploading() tea.Cmd {
	return func() tea.Msg {
		return uploading(true)
	}
}

type bytesCreatedType bool

func bytesCreated() tea.Cmd {
	return func() tea.Msg {
		return bytesCreatedType(true)
	}
}

func initialCreateBytesModel(parentModel model) CreateBytesModel {
	fp := filepicker.New()
	fp.CurrentDirectory, err = os.UserHomeDir()
	fp.Height = 10
	if err != nil {
		panic(err)
	}

	m := CreateBytesModel{
		filepicker:  fp,
		parentModel: &parentModel,
	}

	return m
}

/*
func main() {
	fp := filepicker.New()
	//fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	m := CreateBytesModel{
		filepicker: fp,
	}
	tea.NewProgram(&m).Run()
	//mm := tm.(CreateBytesModel)
	//fmt.Println("\n  You selected: " + m.filepicker.Styles.Selected.Render(mm.selectedFile) + "\n")
}

*/
