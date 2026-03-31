package company

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mkh/rice-railing/internal/constitution"
)

func TestLoadPack(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "company-pack.yaml")

	content := `name: acme-standards
version: "1.0"
description: ACME Corp engineering standards
doctrine:
  architecture:
    target_style: hexagonal
  quality:
    safety_mode: balanced
    block_on:
      - lint
      - typecheck
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing pack file: %v", err)
	}

	pack, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if pack.Name != "acme-standards" {
		t.Errorf("Name: got %q, want %q", pack.Name, "acme-standards")
	}
	if pack.Version != "1.0" {
		t.Errorf("Version: got %q, want %q", pack.Version, "1.0")
	}
	if pack.Description != "ACME Corp engineering standards" {
		t.Errorf("Description: got %q, want %q", pack.Description, "ACME Corp engineering standards")
	}
	if pack.Doctrine.Architecture == nil {
		t.Fatal("Doctrine.Architecture is nil")
	}
	if pack.Doctrine.Architecture.TargetStyle != "hexagonal" {
		t.Errorf("Architecture.TargetStyle: got %q, want %q", pack.Doctrine.Architecture.TargetStyle, "hexagonal")
	}
	if pack.Doctrine.Quality == nil {
		t.Fatal("Doctrine.Quality is nil")
	}
	if pack.Doctrine.Quality.SafetyMode != "balanced" {
		t.Errorf("Quality.SafetyMode: got %q, want %q", pack.Doctrine.Quality.SafetyMode, "balanced")
	}
	if len(pack.Doctrine.Quality.BlockOn) != 2 {
		t.Errorf("Quality.BlockOn: got %d items, want 2", len(pack.Doctrine.Quality.BlockOn))
	}
}

func TestLoadPackMissing(t *testing.T) {
	_, err := Load("/nonexistent/path/pack.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestApplyPack(t *testing.T) {
	pack := &Pack{
		Name:    "test-pack",
		Version: "1.0",
		Doctrine: DoctrinePack{
			Architecture: &constitution.ArchitectureSpec{
				TargetStyle: "hexagonal",
				Modules:     []string{"core", "infra"},
			},
			Quality: &constitution.QualitySpec{
				SafetyMode: "balanced",
				BlockOn:    []string{"lint", "typecheck"},
			},
		},
	}

	c := &constitution.Constitution{}

	Apply(pack, c)

	if c.Architecture.TargetStyle != "hexagonal" {
		t.Errorf("Architecture.TargetStyle: got %q, want %q", c.Architecture.TargetStyle, "hexagonal")
	}
	if len(c.Architecture.Modules) != 2 {
		t.Errorf("Architecture.Modules: got %d, want 2", len(c.Architecture.Modules))
	}
	if c.Quality.SafetyMode != "balanced" {
		t.Errorf("Quality.SafetyMode: got %q, want %q", c.Quality.SafetyMode, "balanced")
	}
	if len(c.Quality.BlockOn) != 2 {
		t.Errorf("Quality.BlockOn: got %d items, want 2", len(c.Quality.BlockOn))
	}
}

func TestApplyPackDoesNotOverride(t *testing.T) {
	pack := &Pack{
		Name:    "test-pack",
		Version: "1.0",
		Doctrine: DoctrinePack{
			Quality: &constitution.QualitySpec{
				SafetyMode: "balanced",
				BlockOn:    []string{"lint"},
			},
			Architecture: &constitution.ArchitectureSpec{
				TargetStyle: "layered",
			},
		},
	}

	c := &constitution.Constitution{
		Quality: constitution.QualitySpec{
			SafetyMode: "strict",
		},
		Architecture: constitution.ArchitectureSpec{
			TargetStyle: "hexagonal",
		},
	}

	Apply(pack, c)

	if c.Quality.SafetyMode != "strict" {
		t.Errorf("Quality.SafetyMode should keep 'strict', got %q", c.Quality.SafetyMode)
	}
	if c.Architecture.TargetStyle != "hexagonal" {
		t.Errorf("Architecture.TargetStyle should keep 'hexagonal', got %q", c.Architecture.TargetStyle)
	}
	// BlockOn was empty in the constitution, so it should be filled from the pack
	if len(c.Quality.BlockOn) != 1 || c.Quality.BlockOn[0] != "lint" {
		t.Errorf("Quality.BlockOn should be filled from pack: got %v", c.Quality.BlockOn)
	}
}
