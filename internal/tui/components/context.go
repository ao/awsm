package components

import (
	"fmt"

	"github.com/ao/awsm/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ContextSwitcher is a component for switching between AWS contexts
type ContextSwitcher struct {
	list         list.Model
	width        int
	height       int
	selectedItem string
	visible      bool
	onSelect     func(string)
}

// contextItem represents a context in the list
type contextItem struct {
	name    string
	profile string
	region  string
	role    string
	current bool
}

// FilterValue implements list.Item interface
func (i contextItem) FilterValue() string {
	return i.name
}

// Title returns the title of the item
func (i contextItem) Title() string {
	if i.current {
		return fmt.Sprintf("* %s", i.name)
	}
	return i.name
}

// Description returns the description of the item
func (i contextItem) Description() string {
	desc := fmt.Sprintf("Profile: %s, Region: %s", i.profile, i.region)
	if i.role != "" {
		desc += fmt.Sprintf(", Role: %s", i.role)
	}
	return desc
}

// NewContextSwitcher creates a new context switcher
func NewContextSwitcher(onSelect func(string)) *ContextSwitcher {
	// Create list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "AWS Contexts"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#9900cc")).
		Padding(0, 1)

	// Custom key bindings
	l.KeyMap.CursorUp = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)
	l.KeyMap.CursorDown = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	)
	l.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	)
	l.KeyMap.ForceQuit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select context"),
			),
		}
	}

	return &ContextSwitcher{
		list:     l,
		visible:  false,
		onSelect: onSelect,
	}
}

// SetSize sets the size of the context switcher
func (c *ContextSwitcher) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.list.SetSize(width, height)
}

// Show shows the context switcher
func (c *ContextSwitcher) Show() {
	c.visible = true
	c.refreshContexts()
}

// Hide hides the context switcher
func (c *ContextSwitcher) Hide() {
	c.visible = false
}

// IsVisible returns whether the context switcher is visible
func (c *ContextSwitcher) IsVisible() bool {
	return c.visible
}

// refreshContexts refreshes the list of contexts
func (c *ContextSwitcher) refreshContexts() {
	// Get contexts
	contexts := config.ListContexts()
	items := make([]list.Item, 0, len(contexts))

	// Add contexts to list
	for _, ctx := range contexts {
		items = append(items, contextItem{
			name:    ctx.Name,
			profile: ctx.Profile,
			region:  ctx.Region,
			role:    ctx.Role,
			current: ctx.Current,
		})
	}

	// Update list
	c.list.SetItems(items)
}

// Init initializes the context switcher
func (c *ContextSwitcher) Init() tea.Cmd {
	return nil
}

// Update handles events for the context switcher
func (c *ContextSwitcher) Update(msg tea.Msg) (*ContextSwitcher, tea.Cmd) {
	if !c.visible {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.list.KeyMap.Quit):
			c.Hide()
			return c, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Get selected item
			if i, ok := c.list.SelectedItem().(contextItem); ok {
				c.selectedItem = i.name
				c.Hide()
				if c.onSelect != nil {
					c.onSelect(i.name)
				}
			}
			return c, nil
		}
	}

	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

// View renders the context switcher
func (c *ContextSwitcher) View() string {
	if !c.visible {
		return ""
	}

	return lipgloss.NewStyle().
		Width(c.width).
		Height(c.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9900cc")).
		Render(c.list.View())
}

// HandleKeyMsg handles key messages for the context switcher
func (c *ContextSwitcher) HandleKeyMsg(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !c.visible {
		// Check if we should show the context switcher
		if msg.String() == "c" {
			c.Show()
			return true, nil
		}
		return false, nil
	}

	// Handle key message
	_, cmd := c.Update(msg)
	return true, cmd
}
