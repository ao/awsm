package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfileSelector(t *testing.T) {
	// Create a profile selector
	onSelect := func(profile string) {
		// Just a placeholder for the callback
	}

	ps := NewProfileSelector(onSelect)

	// Test initial state
	assert.False(t, ps.IsVisible())
	assert.Equal(t, "", ps.selectedItem)

	// Test showing the selector
	ps.Show()
	assert.True(t, ps.IsVisible())

	// Test hiding the selector
	ps.Hide()
	assert.False(t, ps.IsVisible())
}

func TestRegionSelector(t *testing.T) {
	// Create a region selector
	onSelect := func(region string) {
		// Just a placeholder for the callback
	}

	rs := NewRegionSelector(onSelect)

	// Test initial state
	assert.False(t, rs.IsVisible())
	assert.Equal(t, "", rs.selectedItem)

	// Test showing the selector
	rs.Show()
	assert.True(t, rs.IsVisible())

	// Test hiding the selector
	rs.Hide()
	assert.False(t, rs.IsVisible())
}

func TestRegionSelectorContainsAllRegions(t *testing.T) {
	// Create a region selector
	rs := NewRegionSelector(func(string) {})

	// Show the selector to populate the regions
	rs.Show()

	// Get the items from the list
	items := rs.list.Items()

	// Convert items to a map for easier lookup
	regionMap := make(map[string]bool)
	for _, item := range items {
		if ri, ok := item.(regionItem); ok {
			regionMap[ri.name] = true
		}
	}

	// Check for the presence of all AWS regions, including the ones that were missing
	requiredRegions := []string{
		// North America
		"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ca-central-1", "ca-west-1",

		// South America
		"sa-east-1",

		// Europe
		"eu-north-1", "eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"eu-central-2", "eu-south-1", "eu-south-2",

		// Asia Pacific
		"ap-east-1", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
		"ap-southeast-1", "ap-southeast-2", "ap-southeast-3", "ap-southeast-4",
		"ap-south-1", "ap-south-2",

		// Middle East
		"me-south-1", "me-central-1",

		// Africa
		"af-south-1",

		// Israel
		"il-central-1",
	}

	for _, region := range requiredRegions {
		assert.True(t, regionMap[region], "Region %s should be in the list", region)
	}
}

// We can't easily mock the GetAWSProfiles function since it's not a variable
// Instead, we'll test that the ProfileSelector shows items
func TestProfileSelectorShowsItems(t *testing.T) {
	// Create a profile selector
	ps := NewProfileSelector(func(string) {})

	// Show the selector to populate the profiles
	ps.Show()

	// Get the items from the list
	items := ps.list.Items()

	// Verify that we have at least one item (should be at least "default")
	assert.Greater(t, len(items), 0, "ProfileSelector should have at least one profile")
}
