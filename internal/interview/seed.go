package interview

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadSeed reads a seed file from the given path.
func LoadSeed(path string) (*Seed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading seed file %s: %w", path, err)
	}

	var s Seed
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing seed file %s: %w", path, err)
	}

	return &s, nil
}
