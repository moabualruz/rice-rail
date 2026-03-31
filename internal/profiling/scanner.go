package profiling

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// Scanner walks a repository and produces a RepoProfile.
type Scanner struct {
	Root     string
	MaxDepth int
}

// NewScanner creates a scanner for the given repo root.
func NewScanner(root string) *Scanner {
	return &Scanner{
		Root:     root,
		MaxDepth: 5,
	}
}

// Scan performs full repository profiling.
func (s *Scanner) Scan() (*RepoProfile, error) {
	profile := &RepoProfile{}

	langCounts, err := s.detectLanguages()
	if err != nil {
		return nil, fmt.Errorf("detecting languages: %w", err)
	}
	profile.Languages = langCounts

	profile.PackageManagers = s.detectByPatterns(ManifestPatterns)
	profile.BuildSystems = s.detectByPatterns(BuildSystemPatterns)
	profile.CI = s.detectByPatterns(CIPatterns)
	profile.RepoTopology = s.detectTopology()
	profile.Tooling = s.detectTooling()
	profile.ArchHints = s.detectArchHints()

	return profile, nil
}

func (s *Scanner) detectLanguages() ([]DetectedItem, error) {
	counts := map[string]int{}
	total := 0

	err := filepath.WalkDir(s.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".venv" || name == "target" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(s.Root, path)
			depth := strings.Count(rel, string(filepath.Separator))
			if depth >= s.MaxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if lang, ok := LanguagePatterns[ext]; ok {
			counts[lang]++
			total++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var items []DetectedItem
	for lang, count := range counts {
		confidence := float64(count) / float64(max(total, 1))
		if confidence > 1.0 {
			confidence = 1.0
		}
		items = append(items, DetectedItem{
			Name:       lang,
			Confidence: confidence,
			Evidence:   fmt.Sprintf("%d files", count),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Confidence > items[j].Confidence
	})

	return items, nil
}

func (s *Scanner) detectByPatterns(patterns []FilePattern) []DetectedItem {
	seen := map[string]bool{}
	var items []DetectedItem

	for _, p := range patterns {
		matches, _ := filepath.Glob(filepath.Join(s.Root, p.Glob))
		if len(matches) > 0 && !seen[p.Name] {
			seen[p.Name] = true
			rel, _ := filepath.Rel(s.Root, matches[0])
			items = append(items, DetectedItem{
				Name:       p.Name,
				Confidence: 0.95,
				Evidence:   rel,
			})
		}
	}

	return items
}

func (s *Scanner) detectTopology() string {
	for _, p := range MonorepoIndicators {
		matches, _ := filepath.Glob(filepath.Join(s.Root, p.Glob))
		if len(matches) > 0 {
			return "monorepo"
		}
	}

	// Check for Go workspace
	if _, err := os.Stat(filepath.Join(s.Root, "go.work")); err == nil {
		return "monorepo"
	}

	// Check for Cargo workspace
	cargoPath := filepath.Join(s.Root, "Cargo.toml")
	if data, err := os.ReadFile(cargoPath); err == nil {
		if strings.Contains(string(data), "[workspace]") {
			return "monorepo"
		}
	}

	// Check for multiple package.json files (workspaces indicator)
	pkgCount := 0
	filepath.WalkDir(s.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "package.json" {
			pkgCount++
		}
		if pkgCount > 2 {
			return filepath.SkipAll
		}
		return nil
	})
	if pkgCount > 2 {
		return "hybrid"
	}

	return "single"
}

func (s *Scanner) detectTooling() ToolingProfile {
	tp := ToolingProfile{}
	seen := map[string]bool{}

	for _, p := range ToolConfigPatterns {
		matches, _ := filepath.Glob(filepath.Join(s.Root, p.Glob))
		if len(matches) == 0 {
			continue
		}

		// Deduplicate tools by name (e.g., .golangci.yml and .golangci.yaml both map to golangci-lint)
		if seen[p.Name] {
			continue
		}
		seen[p.Name] = true

		binaryName := toolBinaryName(p.Name)
		_, installed := exec.Which(binaryName)
		tool := DetectedTool{
			Name:       p.Name,
			Category:   p.Category,
			ConfigFile: matches[0],
			Installed:  installed,
			Confidence: 0.95,
			Evidence:   fmt.Sprintf("config file: %s", matches[0]),
		}

		switch p.Category {
		case "linter":
			tp.Linters = append(tp.Linters, tool)
		case "formatter":
			tp.Formatters = append(tp.Formatters, tool)
		case "typechecker":
			tp.Typecheckers = append(tp.Typecheckers, tool)
		case "test_runner":
			tp.TestRunners = append(tp.TestRunners, tool)
		case "rule_engine":
			tp.RuleEngines = append(tp.RuleEngines, tool)
		case "codemod":
			tp.Codemods = append(tp.Codemods, tool)
		case "security":
			tp.Security = append(tp.Security, tool)
		}
	}

	return tp
}

// toolBinaryName maps tool names to their actual binary names on PATH.
func toolBinaryName(toolName string) string {
	overrides := map[string]string{
		"typescript":         "tsc",
		"dependency-cruiser": "depcruise",
		"editorconfig":       "editorconfig",
		"pre-commit":         "pre-commit",
		"setuptools":         "python",
		"husky":              "npx",
	}
	if bin, ok := overrides[toolName]; ok {
		return bin
	}
	return toolName
}

func (s *Scanner) detectArchHints() []ArchHint {
	var hints []ArchHint

	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return hints
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if suggestion, ok := ArchitecturePatterns[name]; ok {
			hints = append(hints, ArchHint{
				Pattern:    e.Name() + "/",
				Suggests:   suggestion,
				Confidence: 0.7,
				Evidence:   fmt.Sprintf("directory: %s/", e.Name()),
			})
		}
	}

	return hints
}
