package debug

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// CaptureSnapshot captures the state of an application or component.
// It accepts any interface{} and attempts to extract state information from it.
// If the app implements Snapshottable, it will use that interface.
// Otherwise, it will attempt to extract state using reflection.
func CaptureSnapshot(app interface{}) (*Snapshot, error) {
	snapshot := NewSnapshot()

	// If the app implements MetadataProvider, get metadata
	if provider, ok := app.(MetadataProvider); ok {
		for k, v := range provider.GetSnapshotMetadata() {
			snapshot.AddMetadata(k, v)
		}
	}

	// Add basic metadata
	snapshot.AddMetadata("timestamp", snapshot.Timestamp.Format(time.RFC3339))
	snapshot.AddMetadata("snapshot_id", snapshot.ID)

	// If the app implements Snapshottable, use that interface
	if snapshottable, ok := app.(Snapshottable); ok {
		snapshot.AddState(snapshottable.GetSnapshotID(), snapshottable.GetSnapshotState())
		return snapshot, nil
	}

	// If the app is a struct, try to extract state from its fields
	val := reflect.ValueOf(app)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("app is not a struct or Snapshottable: %T", app)
	}

	// Extract state from struct fields
	err := extractStateFromStruct(snapshot, "", val)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

// extractStateFromStruct recursively extracts state from a struct using reflection.
func extractStateFromStruct(snapshot *Snapshot, prefix string, val reflect.Value) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		fieldName := fieldType.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		// If the field implements Snapshottable, use that interface
		if field.CanInterface() {
			fieldInterface := field.Interface()
			if snapshottable, ok := fieldInterface.(Snapshottable); ok {
				snapshot.AddState(snapshottable.GetSnapshotID(), snapshottable.GetSnapshotState())
				continue
			}
		}

		// Handle different field types
		switch field.Kind() {
		case reflect.Struct:
			// Recursively extract state from nested structs
			if field.CanInterface() {
				err := extractStateFromStruct(snapshot, fieldName, field)
				if err != nil {
					return err
				}
			}
		case reflect.Ptr, reflect.Interface:
			// Handle pointers and interfaces
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				if field.CanInterface() {
					err := extractStateFromStruct(snapshot, fieldName, field.Elem())
					if err != nil {
						return err
					}
				}
			}
		default:
			// For basic types, add them directly if they can be interfaced
			if field.CanInterface() {
				snapshot.AddState(fieldName, field.Interface())
			}
		}
	}

	return nil
}

// GenerateVisualState generates a visual representation of an application or component.
// It accepts any interface{} and attempts to extract visual information from it.
// If the app implements Visualizable, it will use that interface.
// Otherwise, it will attempt to create a basic representation using reflection.
func GenerateVisualState(app interface{}, detailLevel DetailLevel) (*VisualState, error) {
	// Default dimensions
	width, height := 80, 24

	// If the app implements Visualizable, get dimensions from it
	if visualizable, ok := app.(Visualizable); ok {
		width, height = visualizable.GetVisualDimensions()
	}

	visualState := NewVisualState(detailLevel, width, height)

	// If the app implements LayoutProvider, get layout description
	if provider, ok := app.(LayoutProvider); ok {
		visualState.SetLayout(provider.GetLayoutDescription())
	}

	// If the app implements Visualizable, use that interface
	if visualizable, ok := app.(Visualizable); ok {
		visualState.AddComponent(
			visualizable.GetVisualID(),
			visualizable.GetVisualRepresentation(detailLevel),
		)
		return visualState, nil
	}

	// If the app is a struct, try to extract visual information from its fields
	val := reflect.ValueOf(app)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("app is not a struct or Visualizable: %T", app)
	}

	// Generate a basic representation of the struct
	representation := generateStructRepresentation(val, detailLevel)
	visualState.AddComponent("main", representation)

	return visualState, nil
}

// generateStructRepresentation generates a string representation of a struct using reflection.
func generateStructRepresentation(val reflect.Value, detailLevel DetailLevel) string {
	typ := val.Type()
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Type: %s\n", typ.Name()))
	sb.WriteString("Fields:\n")

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		fieldName := fieldType.Name
		fieldKind := field.Kind().String()

		// For detailed view, include more information
		if detailLevel == DetailedDetail {
			if field.CanInterface() {
				fieldValue := field.Interface()
				sb.WriteString(fmt.Sprintf("  %s (%s): %v\n", fieldName, fieldKind, fieldValue))
			} else {
				sb.WriteString(fmt.Sprintf("  %s (%s): <unexported>\n", fieldName, fieldKind))
			}
		} else {
			// For normal and minimal views, just show the field name and type
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", fieldName, fieldKind))
		}

		// For nested structs in detailed view, show more information
		if detailLevel == DetailedDetail && (field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct)) {

			var nestedVal reflect.Value
			if field.Kind() == reflect.Ptr {
				nestedVal = field.Elem()
			} else {
				nestedVal = field
			}

			if field.CanInterface() {
				// Check if it's Visualizable
				if visualizable, ok := field.Interface().(Visualizable); ok {
					sb.WriteString(fmt.Sprintf("    %s\n",
						visualizable.GetVisualRepresentation(detailLevel)))
				} else {
					// Otherwise use reflection
					nestedType := nestedVal.Type()
					sb.WriteString(fmt.Sprintf("    Type: %s\n", nestedType.Name()))

					for j := 0; j < nestedVal.NumField(); j++ {
						nestedField := nestedVal.Field(j)
						nestedFieldType := nestedType.Field(j)

						if nestedFieldType.IsExported() && nestedField.CanInterface() {
							sb.WriteString(fmt.Sprintf("      %s: %v\n",
								nestedFieldType.Name, nestedField.Interface()))
						}
					}
				}
			}
		}
	}

	return sb.String()
}

// FindVisualizableComponents recursively searches for components that implement Visualizable.
// This is useful for automatically discovering components in a complex application structure.
func FindVisualizableComponents(app interface{}) []Visualizable {
	var components []Visualizable

	// If the app is nil, return empty slice
	if app == nil {
		return components
	}

	// If the app itself is Visualizable, add it
	if visualizable, ok := app.(Visualizable); ok {
		components = append(components, visualizable)
	}

	// Use reflection to search for Visualizable components in the app's fields
	val := reflect.ValueOf(app)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return components
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := typ.Field(i)

			// Skip unexported fields
			if !fieldType.IsExported() {
				continue
			}

			if !field.CanInterface() {
				continue
			}

			fieldInterface := field.Interface()

			// If the field implements Visualizable, add it
			if visualizable, ok := fieldInterface.(Visualizable); ok {
				components = append(components, visualizable)
			}

			// Recursively search in struct fields and pointers to structs
			switch field.Kind() {
			case reflect.Struct:
				nestedComponents := FindVisualizableComponents(fieldInterface)
				components = append(components, nestedComponents...)
			case reflect.Ptr, reflect.Interface:
				if !field.IsNil() {
					nestedComponents := FindVisualizableComponents(fieldInterface)
					components = append(components, nestedComponents...)
				}
			}
		}
	}

	return components
}

// FindSnapshottableComponents recursively searches for components that implement Snapshottable.
// This is useful for automatically discovering components in a complex application structure.
func FindSnapshottableComponents(app interface{}) []Snapshottable {
	var components []Snapshottable

	// If the app is nil, return empty slice
	if app == nil {
		return components
	}

	// If the app itself is Snapshottable, add it
	if snapshottable, ok := app.(Snapshottable); ok {
		components = append(components, snapshottable)
	}

	// Use reflection to search for Snapshottable components in the app's fields
	val := reflect.ValueOf(app)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return components
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := typ.Field(i)

			// Skip unexported fields
			if !fieldType.IsExported() {
				continue
			}

			if !field.CanInterface() {
				continue
			}

			fieldInterface := field.Interface()

			// If the field implements Snapshottable, add it
			if snapshottable, ok := fieldInterface.(Snapshottable); ok {
				components = append(components, snapshottable)
			}

			// Recursively search in struct fields and pointers to structs
			switch field.Kind() {
			case reflect.Struct:
				nestedComponents := FindSnapshottableComponents(fieldInterface)
				components = append(components, nestedComponents...)
			case reflect.Ptr, reflect.Interface:
				if !field.IsNil() {
					nestedComponents := FindSnapshottableComponents(fieldInterface)
					components = append(components, nestedComponents...)
				}
			}
		}
	}

	return components
}
