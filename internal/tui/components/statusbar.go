package components

import (
	"fmt"

	"github.com/ao/awsm/internal/config"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar represents the status bar at the bottom of the screen
type StatusBar struct {
	width int
	style lipgloss.Style
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	return &StatusBar{
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1),
	}
}

// SetWidth sets the width of the status bar
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// Render renders the status bar
func (s *StatusBar) Render() string {
	// Get current context, AWS profile and region
	contextName := config.GetCurrentContext()
	profile := config.GetAWSProfile()
	region := config.GetAWSRegion()
	role := config.GetAWSRole()

	// Create status sections
	contextSection := s.style.Copy().
		Background(lipgloss.Color("#9900cc")).
		Render(fmt.Sprintf(" Context: %s ", contextName))

	profileSection := s.style.Copy().
		Background(lipgloss.Color("#0066cc")).
		Render(fmt.Sprintf(" Profile: %s ", profile))

	regionSection := s.style.Copy().
		Background(lipgloss.Color("#006600")).
		Render(fmt.Sprintf(" Region: %s ", region))

	// Help section
	helpSection := s.style.Copy().
		Align(lipgloss.Right).
		Render(" ? for help ")

	// Calculate remaining space
	usedWidth := lipgloss.Width(contextSection) + lipgloss.Width(profileSection) +
		lipgloss.Width(regionSection) + lipgloss.Width(helpSection)

	// Add role section if role is set
	var roleSection string
	if role != "" {
		roleSection = s.style.Copy().
			Background(lipgloss.Color("#cc6600")).
			Render(fmt.Sprintf(" Role: %s ", role))
		usedWidth += lipgloss.Width(roleSection)
	}

	remainingWidth := s.width - usedWidth

	// Create connection status section
	connectionStatus := "Connected"
	statusColor := lipgloss.Color("#006600") // Green for connected

	// In a real implementation, we would check the actual connection status
	// For now, we'll assume we're connected if we have a profile and region
	if profile == "" || region == "" {
		connectionStatus = "Disconnected"
		statusColor = lipgloss.Color("#cc0000") // Red for disconnected
	}

	connectionSection := s.style.Copy().
		Background(statusColor).
		Width(remainingWidth).
		Render(fmt.Sprintf(" Status: %s ", connectionStatus))

	// Combine all sections
	sections := []string{
		contextSection,
		profileSection,
		regionSection,
	}

	// Add role section if role is set
	if role != "" {
		sections = append(sections, roleSection)
	}

	// Add connection and help sections
	sections = append(sections, connectionSection, helpSection)

	return lipgloss.JoinHorizontal(lipgloss.Left, sections...)
}
