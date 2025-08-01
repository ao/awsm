package debug

import (
	"bytes"
	"fmt"
	"strings"
)

// DetailLevel represents the level of detail for visual state representation.
type DetailLevel int

const (
	// MinimalDetail provides a minimal representation with just essential information.
	MinimalDetail DetailLevel = iota
	// NormalDetail provides a standard representation with moderate details.
	NormalDetail
	// DetailedDetail provides a comprehensive representation with all available details.
	DetailedDetail
)

// VisualState represents a text-based visual representation of the TUI application state.
type VisualState struct {
	// Components holds the visual representation of each UI component.
	Components map[string]string
	// Layout describes the overall layout of the components.
	Layout string
	// DetailLevel determines how much detail to include in the representation.
	DetailLevel DetailLevel
	// Width is the width of the visual representation in characters.
	Width int
	// Height is the height of the visual representation in characters.
	Height int
}

// NewVisualState creates a new VisualState with the specified detail level and dimensions.
func NewVisualState(detailLevel DetailLevel, width, height int) *VisualState {
	return &VisualState{
		Components:  make(map[string]string),
		DetailLevel: detailLevel,
		Width:       width,
		Height:      height,
	}
}

// AddComponent adds a visual representation of a UI component.
func (vs *VisualState) AddComponent(name string, representation string) {
	vs.Components[name] = representation
}

// SetLayout sets the overall layout description.
func (vs *VisualState) SetLayout(layout string) {
	vs.Layout = layout
}

// String returns the complete visual representation as a string.
func (vs *VisualState) String() string {
	var buf bytes.Buffer

	// Add a header with dimensions
	buf.WriteString(fmt.Sprintf("Visual State [%dx%d] (Detail: %s)\n",
		vs.Width, vs.Height, vs.detailLevelString()))
	buf.WriteString(strings.Repeat("=", vs.Width) + "\n\n")

	// Add the layout description if available
	if vs.Layout != "" {
		buf.WriteString("Layout: " + vs.Layout + "\n")
		buf.WriteString(strings.Repeat("-", vs.Width) + "\n\n")
	}

	// Add each component
	for name, representation := range vs.Components {
		buf.WriteString(fmt.Sprintf("Component: %s\n", name))
		buf.WriteString(strings.Repeat("-", len(name)+11) + "\n")
		buf.WriteString(representation + "\n\n")
	}

	return buf.String()
}

// detailLevelString returns a string representation of the detail level.
func (vs *VisualState) detailLevelString() string {
	switch vs.DetailLevel {
	case MinimalDetail:
		return "Minimal"
	case NormalDetail:
		return "Normal"
	case DetailedDetail:
		return "Detailed"
	default:
		return "Unknown"
	}
}

// GenerateBoxDrawing creates a box drawing representation of a component.
func GenerateBoxDrawing(title string, content string, width, height int) string {
	var buf bytes.Buffer

	// Ensure title fits within width
	if len(title) > width-4 {
		title = title[:width-4]
	}

	// Top border with title
	topBorder := "┌─" + title + strings.Repeat("─", width-len(title)-3) + "┐"
	// Ensure the top border is exactly width characters
	if len(topBorder) > width {
		topBorder = topBorder[:width]
	} else if len(topBorder) < width {
		// This shouldn't happen with the calculation above, but just in case
		topBorder = topBorder + strings.Repeat("─", width-len(topBorder))
	}
	buf.WriteString(topBorder + "\n")

	// Content lines
	lines := strings.Split(content, "\n")
	for i := 0; i < height-2; i++ {
		line := "│ "
		if i < len(lines) {
			// Truncate or pad the line to fit the width
			contentLine := lines[i]
			if len(contentLine) > width-4 {
				contentLine = contentLine[:width-4]
			}
			line += contentLine + strings.Repeat(" ", width-4-len(contentLine)) + " │"
		} else {
			line += strings.Repeat(" ", width-4) + " │"
		}

		// Ensure the line is exactly width characters
		if len(line) > width {
			line = line[:width]
		} else if len(line) < width {
			// This shouldn't happen with the calculation above, but just in case
			line = line + strings.Repeat(" ", width-len(line))
		}
		buf.WriteString(line + "\n")
	}

	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"
	// Ensure the bottom border is exactly width characters
	if len(bottomBorder) > width {
		bottomBorder = bottomBorder[:width]
	} else if len(bottomBorder) < width {
		// This shouldn't happen with the calculation above, but just in case
		bottomBorder = bottomBorder + strings.Repeat("─", width-len(bottomBorder))
	}
	buf.WriteString(bottomBorder + "\n")

	return buf.String()
}

// GenerateTable creates a table representation of data.
func GenerateTable(headers []string, rows [][]string, width int) string {
	var buf bytes.Buffer

	// Calculate column widths
	colCount := len(headers)
	if colCount == 0 {
		return ""
	}

	colWidths := make([]int, colCount)
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < colCount && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Adjust column widths to fit within the total width
	totalWidth := colCount + 1 // Account for separators
	for _, w := range colWidths {
		totalWidth += w
	}

	if totalWidth > width && colCount > 0 {
		// Scale down columns proportionally
		excess := totalWidth - width
		for i := range colWidths {
			reduction := (excess * colWidths[i]) / totalWidth
			colWidths[i] = max(colWidths[i]-reduction, 3) // Minimum column width of 3
		}
	}

	// Generate the table
	// Header row
	buf.WriteString("┌")
	for i, w := range colWidths {
		buf.WriteString(strings.Repeat("─", w))
		if i < colCount-1 {
			buf.WriteString("┬")
		}
	}
	buf.WriteString("┐\n")

	// Column headers
	buf.WriteString("│")
	for i, header := range headers {
		if i < colCount {
			if len(header) > colWidths[i] {
				header = header[:colWidths[i]]
			}
			buf.WriteString(header + strings.Repeat(" ", colWidths[i]-len(header)) + "│")
		}
	}
	buf.WriteString("\n")

	// Separator
	buf.WriteString("├")
	for i, w := range colWidths {
		buf.WriteString(strings.Repeat("─", w))
		if i < colCount-1 {
			buf.WriteString("┼")
		}
	}
	buf.WriteString("┤\n")

	// Data rows
	for _, row := range rows {
		buf.WriteString("│")
		for i, cell := range row {
			if i < colCount {
				if len(cell) > colWidths[i] {
					cell = cell[:colWidths[i]]
				}
				buf.WriteString(cell + strings.Repeat(" ", colWidths[i]-len(cell)) + "│")
			}
		}
		buf.WriteString("\n")
	}

	// Bottom border
	buf.WriteString("└")
	for i, w := range colWidths {
		buf.WriteString(strings.Repeat("─", w))
		if i < colCount-1 {
			buf.WriteString("┴")
		}
	}
	buf.WriteString("┘\n")

	return buf.String()
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
