package ui

import "github.com/charmbracelet/lipgloss"

var (
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")). // red
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // green

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")) // yellow

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")) // cyan

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")). // purple
			Bold(true)
)
