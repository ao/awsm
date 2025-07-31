package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// OutputFormat defines the available output formats
type OutputFormat string

const (
	// FormatJSON outputs data in JSON format
	FormatJSON OutputFormat = "json"

	// FormatYAML outputs data in YAML format
	FormatYAML OutputFormat = "yaml"

	// FormatTable outputs data in table format
	FormatTable OutputFormat = "table"

	// FormatText outputs data in plain text format
	FormatText OutputFormat = "text"
)

// IsValidOutputFormat checks if the given format is valid
func IsValidOutputFormat(format string) bool {
	switch OutputFormat(format) {
	case FormatJSON, FormatYAML, FormatTable, FormatText:
		return true
	default:
		return false
	}
}

// FormatOutput formats the given data according to the specified format
func FormatOutput(data interface{}, format string) (string, error) {
	switch OutputFormat(format) {
	case FormatJSON:
		return formatJSON(data)
	case FormatYAML:
		return formatYAML(data)
	case FormatTable:
		return formatTable(data)
	case FormatText:
		return formatText(data)
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

// formatJSON formats data as JSON
func formatJSON(data interface{}) (string, error) {
	// Convert data to JSON with pretty formatting
	formatter := prettyjson.NewFormatter()

	output, err := formatter.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error formatting JSON: %w", err)
	}

	return string(output), nil
}

// formatYAML formats data as YAML
func formatYAML(data interface{}) (string, error) {
	// Convert data to YAML
	output, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error formatting YAML: %w", err)
	}

	return string(output), nil
}

// formatTable formats data as a table
func formatTable(data interface{}) (string, error) {
	// Convert data to a slice of maps for table formatting
	var rows []map[string]interface{}

	// Handle different input types
	switch v := data.(type) {
	case []map[string]interface{}:
		rows = v
	case map[string]interface{}:
		rows = []map[string]interface{}{v}
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				rows = append(rows, m)
			} else {
				// Convert item to JSON and then to map
				jsonData, err := json.Marshal(item)
				if err != nil {
					continue
				}

				var m map[string]interface{}
				if err := json.Unmarshal(jsonData, &m); err == nil {
					rows = append(rows, m)
				}
			}
		}
	default:
		// Convert to JSON and then to map
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("error converting data to JSON: %w", err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(jsonData, &m); err != nil {
			return "", fmt.Errorf("error converting JSON to map: %w", err)
		}

		rows = []map[string]interface{}{m}
	}

	if len(rows) == 0 {
		return "No data to display", nil
	}

	// Extract headers from the first row
	var headers []string
	for k := range rows[0] {
		headers = append(headers, k)
	}

	// Create a buffer to store the table output
	buf := new(bytes.Buffer)

	// Create a new table writer
	table := tablewriter.NewWriter(buf)

	// Configure the table with options
	table = tablewriter.NewWriter(buf)
	opts := []tablewriter.Option{
		tablewriter.WithHeader(headers),
	}

	for _, opt := range opts {
		opt(table)
	}

	// Add rows to the table
	for _, row := range rows {
		var values []string
		for _, h := range headers {
			if val, ok := row[h]; ok {
				values = append(values, fmt.Sprintf("%v", val))
			} else {
				values = append(values, "")
			}
		}
		table.Append(values)
	}

	// Render the table
	table.Render()

	return buf.String(), nil
}

// formatText formats data as plain text
func formatText(data interface{}) (string, error) {
	// For simple types, just convert to string
	switch v := data.(type) {
	case string:
		return v, nil
	case []string:
		return strings.Join(v, "\n"), nil
	default:
		// For complex types, use JSON formatting
		return formatJSON(data)
	}
}

// PrintOutput prints the formatted output to stdout
func PrintOutput(data interface{}, format string) error {
	output, err := FormatOutput(data, format)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// PrintError prints an error message to stderr
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
}
