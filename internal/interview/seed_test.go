package interview

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSeed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.yaml")

	content := `answers:
  arch_target: hexagonal
  safety_mode: strict
multi:
  block_on:
    - lint
    - tests
    - typecheck
`
	os.WriteFile(path, []byte(content), 0644)

	seed, err := LoadSeed(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if seed.Answers["arch_target"] != "hexagonal" {
		t.Errorf("expected hexagonal, got %s", seed.Answers["arch_target"])
	}
	if seed.Answers["safety_mode"] != "strict" {
		t.Errorf("expected strict, got %s", seed.Answers["safety_mode"])
	}
	if len(seed.Multi["block_on"]) != 3 {
		t.Errorf("expected 3 block_on items, got %d", len(seed.Multi["block_on"]))
	}
}

func TestLoadSeedMissing(t *testing.T) {
	_, err := LoadSeed("/nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
