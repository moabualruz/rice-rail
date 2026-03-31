package constitution

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultPath = ".project-toolkit/constitution.yaml"

// Load reads a constitution from the given YAML file path.
func Load(path string) (*Constitution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading constitution %s: %w", path, err)
	}

	var c Constitution
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing constitution %s: %w", path, err)
	}

	return &c, nil
}

// Save writes a constitution to the given YAML file path.
func Save(c *Constitution, path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling constitution: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing constitution %s: %w", path, err)
	}

	return nil
}
