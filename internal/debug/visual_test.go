package debug

import (
	"strings"
	"testing"
)

func TestNewVisualState(t *testing.T) {
	vs := NewVisualState(NormalDetail, 80, 24)

	if vs.Components == nil {
		t.Error("Expected non-nil Components")
	}

	if vs.DetailLevel != NormalDetail {
		t.Errorf("Expected DetailLevel to be NormalDetail, got %v", vs.DetailLevel)
	}

	if vs.Width != 80 {
		t.Errorf("Expected Width to be 80, got %d", vs.Width)
	}

	if vs.Height != 24 {
		t.Errorf("Expected Height to be 24, got %d", vs.Height)
	}
}

func TestAddComponent(t *testing.T) {
	vs := NewVisualState(NormalDetail, 80, 24)
	vs.AddComponent("test", "test representation")

	if rep, ok := vs.Components["test"]; !ok {
		t.Error("Expected component to be added")
	} else if rep != "test representation" {
		t.Errorf("Expected representation to be 'test representation', got %s", rep)
	}
}

func TestSetLayout(t *testing.T) {
	vs := NewVisualState(NormalDetail, 80, 24)
	vs.SetLayout("test layout")

	if vs.Layout != "test layout" {
		t.Errorf("Expected Layout to be 'test layout', got %s", vs.Layout)
	}
}

func TestString(t *testing.T) {
	vs := NewVisualState(NormalDetail, 80, 24)
	vs.SetLayout("test layout")
	vs.AddComponent("test", "test representation")

	result := vs.String()

	// Check that the result contains the expected elements
	if !strings.Contains(result, "Visual State [80x24]") {
		t.Error("Expected result to contain dimensions")
	}

	if !strings.Contains(result, "Detail: Normal") {
		t.Error("Expected result to contain detail level")
	}

	if !strings.Contains(result, "Layout: test layout") {
		t.Error("Expected result to contain layout")
	}

	if !strings.Contains(result, "Component: test") {
		t.Error("Expected result to contain component name")
	}

	if !strings.Contains(result, "test representation") {
		t.Error("Expected result to contain component representation")
	}
}

func TestDetailLevelString(t *testing.T) {
	testCases := []struct {
		level    DetailLevel
		expected string
	}{
		{MinimalDetail, "Minimal"},
		{NormalDetail, "Normal"},
		{DetailedDetail, "Detailed"},
		{DetailLevel(99), "Unknown"},
	}

	for _, tc := range testCases {
		vs := NewVisualState(tc.level, 80, 24)
		result := vs.detailLevelString()

		if result != tc.expected {
			t.Errorf("Expected %s for level %d, got %s", tc.expected, tc.level, result)
		}
	}
}

func TestGenerateBoxDrawing(t *testing.T) {
	title := "Test"
	content := "Line 1\nLine 2"
	width := 20
	height := 5

	result := GenerateBoxDrawing(title, content, width, height)

	// Check that the result contains the expected elements
	if !strings.Contains(result, "┌─Test") {
		t.Error("Expected result to contain title in top border")
	}

	if !strings.Contains(result, "│ Line 1") {
		t.Error("Expected result to contain first line of content")
	}

	if !strings.Contains(result, "│ Line 2") {
		t.Error("Expected result to contain second line of content")
	}

	if !strings.Contains(result, "└") {
		t.Error("Expected result to contain bottom-left corner")
	}

	// Check dimensions
	lines := strings.Split(result, "\n")
	if len(lines) != height+1 { // +1 because Split includes an empty string after the last newline
		t.Errorf("Expected %d lines, got %d", height+1, len(lines))
	}

	for i, line := range lines {
		if i < height && len(line) != width {
			t.Errorf("Expected line %d to have width %d, got %d", i, width, len(line))
		}
	}
}

func TestGenerateTable(t *testing.T) {
	headers := []string{"Col1", "Col2"}
	rows := [][]string{
		{"A", "B"},
		{"C", "D"},
	}
	width := 20

	result := GenerateTable(headers, rows, width)

	// Check that the result contains the expected elements
	if !strings.Contains(result, "Col1") {
		t.Error("Expected result to contain first header")
	}

	if !strings.Contains(result, "Col2") {
		t.Error("Expected result to contain second header")
	}

	if !strings.Contains(result, "A") {
		t.Error("Expected result to contain first cell")
	}

	if !strings.Contains(result, "B") {
		t.Error("Expected result to contain second cell")
	}

	if !strings.Contains(result, "C") {
		t.Error("Expected result to contain third cell")
	}

	if !strings.Contains(result, "D") {
		t.Error("Expected result to contain fourth cell")
	}

	// Check table structure
	if !strings.Contains(result, "┌") {
		t.Error("Expected result to contain top-left corner")
	}

	if !strings.Contains(result, "┬") {
		t.Error("Expected result to contain top separator")
	}

	if !strings.Contains(result, "├") {
		t.Error("Expected result to contain left separator")
	}

	if !strings.Contains(result, "┼") {
		t.Error("Expected result to contain middle separator")
	}

	if !strings.Contains(result, "└") {
		t.Error("Expected result to contain bottom-left corner")
	}

	if !strings.Contains(result, "┴") {
		t.Error("Expected result to contain bottom separator")
	}
}

func TestMax(t *testing.T) {
	testCases := []struct {
		a, b, expected int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{0, 0, 0},
		{-1, -2, -1},
	}

	for _, tc := range testCases {
		result := max(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("Expected max(%d, %d) to be %d, got %d", tc.a, tc.b, tc.expected, result)
		}
	}
}
