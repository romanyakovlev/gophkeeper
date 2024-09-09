package main

import (
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 30

var docStyle = lipgloss.NewStyle().Margin(2)

type item struct {
	ID     string
	Name   string
	UserID string
	Type   string
}

func (i item) Title() string       { return i.Name }
func (i item) Description() string { return "ID:" + i.ID + "; Type:" + i.Type }
func (i item) FilterValue() string { return i.Name }
