package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/romanyakovlev/gophkeeper/internal/client"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"strings"
)

type CreateCreditCardModel struct {
	inputs      []textinput.Model
	focused     int
	err         error
	parentModel *model
}

type creditCardView bool

func createCreditCard() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return creditCardView(true)
	}
}

type creditCardCreatedType bool

func creditCardCreated() bubbletea.Cmd {
	return func() bubbletea.Msg {
		return creditCardCreatedType(true)
	}
}

func (m CreateCreditCardModel) handleAction() bubbletea.Cmd {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	keeper := client.NewKeeperServiceClient(conn)

	_, err = keeper.SaveCreditCard(context.Background(), m.inputs[ccn].Value(), m.inputs[exp].Value(), m.inputs[cvv].Value())
	if err != nil {
		return logToUI(fmt.Sprintf("Failed to create: %v", err))
	}

	return creditCardCreated()
}

func (m *CreateCreditCardModel) updateInputs(msg bubbletea.Msg) bubbletea.Cmd {
	cmds := make([]bubbletea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return bubbletea.Batch(cmds...)
}

type (
	errMsg error
)

const (
	ccn = iota
	exp
	cvv
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

// Validator functions to ensure valid input
func ccnValidator(s string) error {
	// Credit Card Number should a string less than 20 digits
	// It should include 16 integers and 3 spaces
	if len(s) > 16+3 {
		return fmt.Errorf("CCN is too long")
	}

	if len(s) == 0 || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("CCN is invalid")
	}

	// The last digit should be a number unless it is a multiple of 4 in which
	// case it should be a space
	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("CCN must separate groups with spaces")
	}

	// The remaining digits should be integers
	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)

	return err
}

func expValidator(s string) error {
	// The 3 character should be a slash (/)
	// The rest should be numbers
	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fmt.Errorf("EXP is invalid")
	}

	// There should be only one slash and it should be in the 2nd index (3rd character)
	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP is invalid")
	}

	return nil
}

func cvvValidator(s string) error {
	// The CVV should be a number of 3 digits
	// Since the input will already ensure that the CVV is a string of length 3,
	// All we need to do is check that it is a number
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}

func initialCreateCreditCardModel(parentModel model) CreateCreditCardModel {
	var inputs []textinput.Model = make([]textinput.Model, 3)
	inputs[ccn] = textinput.New()
	inputs[ccn].Placeholder = "4505 **** **** 1234"
	inputs[ccn].Focus()
	inputs[ccn].CharLimit = 20
	inputs[ccn].Width = 30
	inputs[ccn].Prompt = ""
	inputs[ccn].Validate = ccnValidator

	inputs[exp] = textinput.New()
	inputs[exp].Placeholder = "MM/YY "
	inputs[exp].CharLimit = 5
	inputs[exp].Width = 5
	inputs[exp].Prompt = ""
	inputs[exp].Validate = expValidator

	inputs[cvv] = textinput.New()
	inputs[cvv].Placeholder = "XXX"
	inputs[cvv].CharLimit = 3
	inputs[cvv].Width = 5
	inputs[cvv].Prompt = ""
	inputs[cvv].Validate = cvvValidator

	return CreateCreditCardModel{
		inputs:      inputs,
		focused:     0,
		err:         nil,
		parentModel: &parentModel,
	}
}

func (m CreateCreditCardModel) Init() bubbletea.Cmd {
	return textinput.Blink
}

func (m CreateCreditCardModel) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	var cmds []bubbletea.Cmd = make([]bubbletea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case creditCardCreatedType:
		return m.parentModel, creditCardCreated()
	case bubbletea.KeyMsg:
		switch msg.Type {
		case bubbletea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				return m, m.handleAction()
			}
			m.nextInput()
		case bubbletea.KeyCtrlC:
			return m, bubbletea.Quit
		case bubbletea.KeyEsc:
			return m.parentModel, m.parentModel.Init()
		case bubbletea.KeyShiftTab, bubbletea.KeyCtrlP:
			m.prevInput()
		case bubbletea.KeyTab, bubbletea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil

	case logMsg:
		return m.parentModel, logToUI(string(msg))
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, bubbletea.Batch(cmds...)
}

func (m CreateCreditCardModel) View() string {
	return fmt.Sprintf(
		`

 %s
 %s

 %s  %s
 %s  %s

 %s
`,
		inputStyle.Width(30).Render("Card Number"),
		m.inputs[ccn].View(),
		inputStyle.Width(6).Render("EXP"),
		inputStyle.Width(6).Render("CVV"),
		m.inputs[exp].View(),
		m.inputs[cvv].View(),
		continueStyle.Render("Continue ->"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *CreateCreditCardModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *CreateCreditCardModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
