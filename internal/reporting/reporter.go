package reporting

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Format controls output format.
type Format int

const (
	FormatText Format = iota
	FormatJSON
	FormatYAML
)

// Reporter writes structured output to a destination.
type Reporter struct {
	Format Format
	Writer io.Writer
}

// New creates a reporter that writes to stdout.
func New(format Format) *Reporter {
	return &Reporter{
		Format: format,
		Writer: os.Stdout,
	}
}

// Print outputs data in the configured format.
func (r *Reporter) Print(data any) error {
	switch r.Format {
	case FormatJSON:
		enc := json.NewEncoder(r.Writer)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case FormatYAML:
		return yaml.NewEncoder(r.Writer).Encode(data)
	default:
		_, err := fmt.Fprintf(r.Writer, "%v\n", data)
		return err
	}
}

// Section prints a labeled section header (text mode only).
func (r *Reporter) Section(title string) {
	if r.Format == FormatText {
		fmt.Fprintf(r.Writer, "\n=== %s ===\n\n", title)
	}
}

// Item prints a key-value pair (text mode only).
func (r *Reporter) Item(key, value string) {
	if r.Format == FormatText {
		fmt.Fprintf(r.Writer, "  %-24s %s\n", key+":", value)
	}
}

// Status prints a status line with a label and state.
func (r *Reporter) Status(label, state string) {
	if r.Format == FormatText {
		fmt.Fprintf(r.Writer, "  [%s] %s\n", state, label)
	}
}

// WriteFile writes structured data to a file path as YAML.
func WriteFile(path string, data any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling output: %w", err)
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
