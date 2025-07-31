package components

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResultsPanel represents a panel for displaying results
type ResultsPanel struct {
	width            int
	height           int
	title            string
	content          string
	loading          bool
	error            error
	style            lipgloss.Style
	borderStyle      lipgloss.Style
	titleStyle       lipgloss.Style
	loadingStartTime time.Time
	loadingTimeout   time.Duration
}

// NewResultsPanel creates a new results panel
func NewResultsPanel() *ResultsPanel {
	return &ResultsPanel{
		width:          80,
		height:         24,
		title:          "Results",
		content:        "",
		loading:        false,
		error:          nil,
		loadingTimeout: 30 * time.Second, // Default timeout of 30 seconds
		style: lipgloss.NewStyle().
			Padding(1, 2),
		borderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0066cc")).
			BorderStyle(lipgloss.RoundedBorder()),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0066cc")).
			Padding(0, 1),
	}
}

// SetSize sets the size of the panel
func (p *ResultsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetTitle sets the title of the panel
func (p *ResultsPanel) SetTitle(title string) {
	p.title = title
}

// SetContent sets the content of the panel
func (p *ResultsPanel) SetContent(content string) {
	p.content = content
	p.loading = false
	p.error = nil
}

// SetLoading sets the loading state of the panel
func (p *ResultsPanel) SetLoading(loading bool) {
	p.loading = loading
	if loading {
		p.loadingStartTime = time.Now()
	}
}

// SetError sets the error state of the panel
func (p *ResultsPanel) SetError(err error) {
	p.error = err
	p.loading = false
}

// SetLoadingTimeout sets the timeout duration for loading states
func (p *ResultsPanel) SetLoadingTimeout(timeout time.Duration) {
	p.loadingTimeout = timeout
}

// CheckTimeout checks if the loading has timed out and returns a command if it has
func (p *ResultsPanel) CheckTimeout() tea.Cmd {
	if !p.loading || p.loadingStartTime.IsZero() {
		return nil
	}

	if time.Since(p.loadingStartTime) > p.loadingTimeout {
		return func() tea.Msg {
			return TimeoutMsg{
				Message: "Operation timed out",
			}
		}
	}

	return nil
}

// TimeoutMsg is a message sent when a loading operation times out
type TimeoutMsg struct {
	Message string
}

// Render renders the panel
func (p *ResultsPanel) Render() string {
	// Calculate available space for content
	availableHeight := p.height - 4 // Account for borders and title

	// Prepare the content
	var displayContent string
	if p.loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true)
		displayContent = loadingStyle.Render("Loading...")

		// Add timeout information if loading for more than 5 seconds
		if !p.loadingStartTime.IsZero() && time.Since(p.loadingStartTime) > 5*time.Second {
			elapsed := time.Since(p.loadingStartTime).Round(time.Second)
			displayContent += "\n\n" + loadingStyle.Render("Operation running for "+elapsed.String())
		}
	} else if p.error != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
		displayContent = errorStyle.Render("Error: " + p.error.Error())
	} else if p.content == "" {
		displayContent = lipgloss.NewStyle().
			Faint(true).
			Render("No data available")
	} else {
		displayContent = p.content
	}

	// Style the content
	styledContent := p.style.Copy().
		Width(p.width - 4). // Account for border and padding
		Height(availableHeight).
		Render(displayContent)

	// Create the title
	styledTitle := p.titleStyle.Copy().Render(" " + p.title + " ")

	// Create the panel with border
	panel := p.borderStyle.Copy().
		Width(p.width).
		Height(p.height).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		Render(styledContent)

	// Create a title row with the title centered
	titleWidth := lipgloss.Width(styledTitle)
	leftPadding := (p.width - titleWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Create padding on both sides of the title
	leftPad := strings.Repeat(" ", leftPadding)
	rightPad := strings.Repeat(" ", p.width-leftPadding-titleWidth)

	// Create a title row with the title centered
	titleRow := leftPad + styledTitle + rightPad

	// Split the panel into lines
	panelLines := strings.Split(panel, "\n")

	// Replace the first line with our title row
	if len(panelLines) > 0 {
		panelLines[0] = titleRow
	}

	// Join the lines back together
	return strings.Join(panelLines, "\n")
}
