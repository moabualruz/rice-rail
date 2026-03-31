package provenance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Record tracks why an artifact or decision was created.
type Record struct {
	ID        string    `yaml:"id" json:"id"`
	Type      string    `yaml:"type" json:"type"` // inferred, user_stated, generated, company_pack
	Source    string    `yaml:"source" json:"source"`
	Artifact  string    `yaml:"artifact,omitempty" json:"artifact,omitempty"`
	Rationale string    `yaml:"rationale" json:"rationale"`
	Agent     string    `yaml:"agent,omitempty" json:"agent,omitempty"`
	Timestamp time.Time `yaml:"timestamp" json:"timestamp"`
}

// Log is an append-only list of provenance records.
type Log struct {
	Records []Record `yaml:"records" json:"records"`
}

// Append adds a record to the log.
func (l *Log) Append(r Record) {
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	l.Records = append(l.Records, r)
}

// Save writes the log to a JSON file at the given path.
func Save(l *Log, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating provenance dir: %w", err)
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling provenance log: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing provenance log: %w", err)
	}
	return nil
}

// Load reads a provenance log from a JSON file.
func Load(path string) (*Log, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading provenance log: %w", err)
	}
	var l Log
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("parsing provenance log: %w", err)
	}
	return &l, nil
}
