package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap is a wrapper around a slice of key bindings that implements help.KeyMap
type KeyMap struct {
	bindings []key.Binding
}

// NewKeyMap creates a new key map
func NewKeyMap(bindings ...key.Binding) KeyMap {
	return KeyMap{bindings: bindings}
}

// ShortHelp returns the short help for the key map
func (k KeyMap) ShortHelp() []key.Binding {
	return k.bindings
}

// FullHelp returns the full help for the key map
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.bindings}
}

// SectionedKeyMap is a key map with named sections
type SectionedKeyMap struct {
	sections map[string][]key.Binding
}

// NewSectionedKeyMap creates a new sectioned key map
func NewSectionedKeyMap() *SectionedKeyMap {
	return &SectionedKeyMap{
		sections: make(map[string][]key.Binding),
	}
}

// AddSection adds a section to the key map
func (s *SectionedKeyMap) AddSection(name string, bindings ...key.Binding) {
	s.sections[name] = bindings
}

// ShortHelp returns the short help for the key map
func (s SectionedKeyMap) ShortHelp() []key.Binding {
	var bindings []key.Binding
	for _, sectionBindings := range s.sections {
		bindings = append(bindings, sectionBindings...)
	}
	return bindings
}

// FullHelp returns the full help for the key map
func (s SectionedKeyMap) FullHelp() [][]key.Binding {
	var sections [][]key.Binding
	for _, sectionBindings := range s.sections {
		sections = append(sections, sectionBindings)
	}
	return sections
}

// HelpView represents a help view component
type HelpView struct {
	help   help.Model
	width  int
	height int
	style  lipgloss.Style
	modal  bool // Whether to render as a modal dialog
}

// NewHelpView creates a new help view
func NewHelpView() *HelpView {
	helpModel := help.New()
	helpModel.ShowAll = false

	return &HelpView{
		help: helpModel,
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(1, 2),
		modal: true, // Default to modal display
	}
}

// SetSize sets the size of the help view
func (h *HelpView) SetSize(width, height int) {
	h.width = width
	h.height = height
	h.help.Width = width - 4 // Account for padding
}

// SetShowAll sets whether to show all help items
func (h *HelpView) SetShowAll(showAll bool) {
	h.help.ShowAll = showAll
}

// ToggleShowAll toggles whether to show all help items
func (h *HelpView) ToggleShowAll() {
	h.help.ShowAll = !h.help.ShowAll
}

// SetModal sets whether to render as a modal dialog
func (h *HelpView) SetModal(modal bool) {
	h.modal = modal
}

// Render renders the help view with the given key map
func (h *HelpView) Render(keyMap help.KeyMap) string {
	helpText := h.help.View(keyMap)

	// Create a title for the help view
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0066cc")).
		Padding(0, 1).
		Render(" Keyboard Shortcuts ")

	// Create a styled help view with a border
	helpView := h.style.Copy().
		Width(h.width - 4). // Account for border width
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Render(helpText)

	// Combine title and help view
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		helpView,
	)

	if !h.modal {
		return content
	}

	// For modal display, create a centered dialog box
	modalWidth := h.width * 3 / 4
	if modalWidth > 100 {
		modalWidth = 100
	} else if modalWidth < 60 {
		modalWidth = 60
	}

	modalHeight := h.height * 2 / 3
	if modalHeight > 30 {
		modalHeight = 30
	} else if modalHeight < 10 {
		modalHeight = 10
	}

	// Create a modal dialog style
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0066cc")).
		Background(lipgloss.Color("#333333")).
		Align(lipgloss.Center).
		Padding(1, 2)

	// Render the modal content
	modalContent := modalStyle.Render(content)

	// Calculate position to center the modal
	leftPadding := (h.width - lipgloss.Width(modalContent)) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	topPadding := (h.height - lipgloss.Height(modalContent)) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Create a full-screen container with the modal centered
	// Use a semi-transparent background to show content behind
	containerStyle := lipgloss.NewStyle().
		Width(h.width).
		Height(h.height).
		Background(lipgloss.Color("rgba(0,0,0,0.7)")).
		Padding(topPadding, leftPadding)

	return containerStyle.Render(modalContent)
}

// RenderBindings renders a help view with the given bindings
func (h *HelpView) RenderBindings(bindings ...key.Binding) string {
	keyMap := NewKeyMap(bindings...)
	return h.Render(keyMap)
}

// RenderSections renders a help view with the given sections
func (h *HelpView) RenderSections(sections map[string][]key.Binding) string {
	keyMap := NewSectionedKeyMap()
	for name, bindings := range sections {
		keyMap.AddSection(name, bindings...)
	}
	return h.Render(keyMap)
}
