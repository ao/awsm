package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Logo represents the ASCII art logo component
type Logo struct {
	width  int
	height int
	style  lipgloss.Style
}

// NewLogo creates a new logo component
func NewLogo() *Logo {
	return &Logo{
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9900")). // AWS Orange color
			Bold(true),
		width:  25, // Default width
		height: 5,  // Default height (5 lines)
	}
}

// SetSize sets the size of the logo
func (l *Logo) SetSize(width, height int) {
	// Ensure the logo has enough space but doesn't take too much
	// We'll use at most 1/4 of the screen width, but at least 25 characters
	logoWidth := width / 4
	if logoWidth < 25 {
		logoWidth = 25 // Minimum width to display the logo properly
	} else if logoWidth > 40 {
		logoWidth = 40 // Maximum width to prevent the logo from being too large
	}

	l.width = logoWidth
	l.height = height
}

// Render renders the logo
func (l *Logo) Render() string {
	// Enhanced ASCII art for "awsm" (AWS CLI Made Awesome)
	// More professional and visually striking design with solid appearance
	logoLines := []string{
		"                              ",
		"    ___        ______  __  __ ",
		"   / \" \"      / / ___||  \"/  |",
		"  / _ \" \" /\" / /\"___ \"| |\"/| |",
		" / ___ \" V  V /  ___) | |  | |",
		"/_/   \"_\"_/\"_/  |____/|_|  |_|",
	}

	// Apply different colors to different parts of the logo
	coloredLogo := []string{}

	// AWS Orange for the first line
	coloredLogo = append(coloredLogo, l.style.Copy().
		Foreground(lipgloss.Color("#FF9900")).
		Bold(true).
		Render(logoLines[0]))

	// AWS Blue for the second line
	coloredLogo = append(coloredLogo, l.style.Copy().
		Foreground(lipgloss.Color("#232F3E")).
		Bold(true).
		Render(logoLines[1]))

	// AWS Orange for the third line
	coloredLogo = append(coloredLogo, l.style.Copy().
		Foreground(lipgloss.Color("#FF9900")).
		Bold(true).
		Render(logoLines[2]))

	// AWS Blue for the fourth line
	coloredLogo = append(coloredLogo, l.style.Copy().
		Foreground(lipgloss.Color("#232F3E")).
		Bold(true).
		Render(logoLines[3]))

	// AWS Orange for the fifth line
	coloredLogo = append(coloredLogo, l.style.Copy().
		Foreground(lipgloss.Color("#FF9900")).
		Bold(true).
		Render(logoLines[4]))

	// Join the colored lines
	logo := lipgloss.JoinVertical(lipgloss.Left, coloredLogo...)

	// Create a container for the logo with appropriate styling
	// Add top padding to prevent the logo from being cut off
	container := lipgloss.NewStyle().
		Padding(1, 1, 0, 1). // Top, right, bottom, left padding
		Align(lipgloss.Right).
		Width(l.width).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#232F3E")).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	// If terminal is too small, return a simplified version
	if l.width < 20 {
		return container.Render("awsm")
	}

	return container.Render(logo)
}
