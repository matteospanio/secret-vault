package tui

import "github.com/charmbracelet/lipgloss"

// Color palette - minimal design system
var (
	colorPrimary   = lipgloss.Color("12")  // Blue
	colorSecondary = lipgloss.Color("8")   // Gray
	colorWarning   = lipgloss.Color("11")  // Yellow
	colorError     = lipgloss.Color("9")   // Red
	colorSuccess   = lipgloss.Color("10")  // Green
	colorDim       = lipgloss.Color("240") // Dim gray
)

// Text styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)
)

// Component styles
var (
	listItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	listSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(colorPrimary).
				Bold(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorDim)

	footerStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorDim)
)
