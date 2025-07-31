package components

import (
	"fmt"

	"github.com/ao/awsm/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProfileSelector is a component for selecting AWS profiles
type ProfileSelector struct {
	list         list.Model
	width        int
	height       int
	selectedItem string
	visible      bool
	onSelect     func(string)
}

// profileItem represents a profile in the list
type profileItem struct {
	name    string
	current bool
}

// FilterValue implements list.Item interface
func (i profileItem) FilterValue() string {
	return i.name
}

// Title returns the title of the item
func (i profileItem) Title() string {
	if i.current {
		return fmt.Sprintf("* %s", i.name)
	}
	return i.name
}

// Description returns the description of the item
func (i profileItem) Description() string {
	return ""
}

// NewProfileSelector creates a new profile selector
func NewProfileSelector(onSelect func(string)) *ProfileSelector {
	// Create list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "AWS Profiles"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0066cc")).
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
				key.WithHelp("enter", "select profile"),
			),
		}
	}

	return &ProfileSelector{
		list:     l,
		visible:  false,
		onSelect: onSelect,
	}
}

// SetSize sets the size of the profile selector
func (p *ProfileSelector) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.list.SetSize(width, height)
}

// Show shows the profile selector
func (p *ProfileSelector) Show() {
	p.visible = true
	p.refreshProfiles()
}

// Hide hides the profile selector
func (p *ProfileSelector) Hide() {
	p.visible = false
}

// IsVisible returns whether the profile selector is visible
func (p *ProfileSelector) IsVisible() bool {
	return p.visible
}

// refreshProfiles refreshes the list of profiles
func (p *ProfileSelector) refreshProfiles() {
	// Get current profile
	currentProfile := config.GetAWSProfile()

	// Get all available profiles
	allProfiles, err := config.GetAWSProfiles()
	if err != nil {
		// If there's an error, fall back to recent profiles
		allProfiles = config.GlobalConfig.Recent.Profiles
	}

	// Add profiles to list
	items := make([]list.Item, 0, len(allProfiles))
	for _, profile := range allProfiles {
		items = append(items, profileItem{
			name:    profile,
			current: profile == currentProfile,
		})
	}

	// Update list
	p.list.SetItems(items)
}

// Init initializes the profile selector
func (p *ProfileSelector) Init() tea.Cmd {
	return nil
}

// Update handles events for the profile selector
func (p *ProfileSelector) Update(msg tea.Msg) (*ProfileSelector, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.list.KeyMap.Quit):
			p.Hide()
			return p, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Get selected item
			if i, ok := p.list.SelectedItem().(profileItem); ok {
				p.selectedItem = i.name
				p.Hide()
				if p.onSelect != nil {
					p.onSelect(i.name)
				}
			}
			return p, nil
		}
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

// View renders the profile selector
func (p *ProfileSelector) View() string {
	if !p.visible {
		return ""
	}

	return lipgloss.NewStyle().
		Width(p.width).
		Height(p.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0066cc")).
		Render(p.list.View())
}

// HandleKeyMsg handles key messages for the profile selector
func (p *ProfileSelector) HandleKeyMsg(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !p.visible {
		// Check if we should show the profile selector
		if msg.String() == "p" {
			p.Show()
			return true, nil
		}
		return false, nil
	}

	// Handle key message
	_, cmd := p.Update(msg)
	return true, cmd
}

// RegionSelector is a component for selecting AWS regions
type RegionSelector struct {
	list         list.Model
	width        int
	height       int
	selectedItem string
	visible      bool
	onSelect     func(string)
}

// regionItem represents a region in the list
type regionItem struct {
	name    string
	current bool
}

// FilterValue implements list.Item interface
func (i regionItem) FilterValue() string {
	return i.name
}

// Title returns the title of the item
func (i regionItem) Title() string {
	if i.current {
		return fmt.Sprintf("* %s", i.name)
	}
	return i.name
}

// Description returns the description of the item
func (i regionItem) Description() string {
	return ""
}

// NewRegionSelector creates a new region selector
func NewRegionSelector(onSelect func(string)) *RegionSelector {
	// Create list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "AWS Regions"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#00cc66")).
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
				key.WithHelp("enter", "select region"),
			),
		}
	}

	return &RegionSelector{
		list:     l,
		visible:  false,
		onSelect: onSelect,
	}
}

// SetSize sets the size of the region selector
func (r *RegionSelector) SetSize(width, height int) {
	r.width = width
	r.height = height
	r.list.SetSize(width, height)
}

// Show shows the region selector
func (r *RegionSelector) Show() {
	r.visible = true
	r.refreshRegions()
}

// Hide hides the region selector
func (r *RegionSelector) Hide() {
	r.visible = false
}

// IsVisible returns whether the region selector is visible
func (r *RegionSelector) IsVisible() bool {
	return r.visible
}

// refreshRegions refreshes the list of regions
func (r *RegionSelector) refreshRegions() {
	// Get current region
	currentRegion := config.GetAWSRegion()

	// All AWS regions
	allRegions := []string{
		// North America
		"us-east-1",    // US East (N. Virginia)
		"us-east-2",    // US East (Ohio)
		"us-west-1",    // US West (N. California)
		"us-west-2",    // US West (Oregon)
		"ca-central-1", // Canada (Central)
		"ca-west-1",    // Canada West (Calgary)

		// South America
		"sa-east-1", // South America (São Paulo)

		// Europe
		"eu-north-1",   // Europe (Stockholm)
		"eu-west-1",    // Europe (Ireland)
		"eu-west-2",    // Europe (London)
		"eu-west-3",    // Europe (Paris)
		"eu-central-1", // Europe (Frankfurt)
		"eu-central-2", // Europe (Zurich)
		"eu-south-1",   // Europe (Milan)
		"eu-south-2",   // Europe (Spain)

		// Asia Pacific
		"ap-east-1",      // Asia Pacific (Hong Kong)
		"ap-northeast-1", // Asia Pacific (Tokyo)
		"ap-northeast-2", // Asia Pacific (Seoul)
		"ap-northeast-3", // Asia Pacific (Osaka)
		"ap-southeast-1", // Asia Pacific (Singapore)
		"ap-southeast-2", // Asia Pacific (Sydney)
		"ap-southeast-3", // Asia Pacific (Jakarta)
		"ap-southeast-4", // Asia Pacific (Melbourne)
		"ap-south-1",     // Asia Pacific (Mumbai)
		"ap-south-2",     // Asia Pacific (Hyderabad)

		// Middle East
		"me-south-1",   // Middle East (Bahrain)
		"me-central-1", // Middle East (UAE)

		// Africa
		"af-south-1", // Africa (Cape Town)

		// China
		"cn-north-1",     // China (Beijing)
		"cn-northwest-1", // China (Ningxia)

		// AWS GovCloud
		"us-gov-east-1", // AWS GovCloud (US-East)
		"us-gov-west-1", // AWS GovCloud (US-West)

		// Israel
		"il-central-1", // Israel (Tel Aviv)
	}

	// Get recent regions
	recentRegions := config.GlobalConfig.Recent.Regions

	// Combine and deduplicate regions
	regionMap := make(map[string]bool)
	for _, region := range recentRegions {
		regionMap[region] = true
	}
	for _, region := range allRegions {
		regionMap[region] = true
	}

	// Create list items
	items := make([]list.Item, 0, len(regionMap))
	for region := range regionMap {
		items = append(items, regionItem{
			name:    region,
			current: region == currentRegion,
		})
	}

	// Update list
	r.list.SetItems(items)
}

// Init initializes the region selector
func (r *RegionSelector) Init() tea.Cmd {
	return nil
}

// Update handles events for the region selector
func (r *RegionSelector) Update(msg tea.Msg) (*RegionSelector, tea.Cmd) {
	if !r.visible {
		return r, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, r.list.KeyMap.Quit):
			r.Hide()
			return r, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Get selected item
			if i, ok := r.list.SelectedItem().(regionItem); ok {
				r.selectedItem = i.name
				r.Hide()
				if r.onSelect != nil {
					r.onSelect(i.name)
				}
			}
			return r, nil
		}
	}

	var cmd tea.Cmd
	r.list, cmd = r.list.Update(msg)
	return r, cmd
}

// View renders the region selector
func (r *RegionSelector) View() string {
	if !r.visible {
		return ""
	}

	return lipgloss.NewStyle().
		Width(r.width).
		Height(r.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00cc66")).
		Render(r.list.View())
}

// HandleKeyMsg handles key messages for the region selector
func (r *RegionSelector) HandleKeyMsg(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !r.visible {
		// Check if we should show the region selector
		if msg.String() == "r" {
			r.Show()
			return true, nil
		}
		return false, nil
	}

	// Handle key message
	_, cmd := r.Update(msg)
	return true, cmd
}
