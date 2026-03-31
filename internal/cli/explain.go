package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/reporting"
)

var explainCmd = &cobra.Command{
	Use:   "explain [rule-or-artifact]",
	Short: "Explain why a rule, tool, or action exists",
	Long:  "Show origin, purpose, enforcement path, and waiver options for any rule or generated artifact.",
	Args:  cobra.ExactArgs(1),
	RunE:  runExplain,
}

func runExplain(cmd *cobra.Command, args []string) error {
	target := args[0]
	rep := reporting.New(getFormat())

	rep.Section(fmt.Sprintf("Explaining: %s", target))

	// Check if it's a file in the toolkit
	toolkitPath := filepath.Join(config.ToolkitDir, target)
	if _, err := os.Stat(toolkitPath); err == nil {
		data, err := os.ReadFile(toolkitPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", toolkitPath, err)
		}
		rep.Item("Type", "toolkit artifact")
		rep.Item("Path", toolkitPath)
		fmt.Println(string(data))
		return nil
	}

	// Check rules directories
	ruleDirs := []string{
		filepath.Join(config.RulesDir, "semgrep"),
		filepath.Join(config.RulesDir, "ast-grep"),
		filepath.Join(config.RulesDir, "custom"),
	}

	for _, dir := range ruleDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
			if name == target || e.Name() == target {
				rulePath := filepath.Join(dir, e.Name())
				data, err := os.ReadFile(rulePath)
				if err != nil {
					return fmt.Errorf("reading rule %s: %w", rulePath, err)
				}
				rep.Item("Type", "rule")
				rep.Item("Engine", filepath.Base(dir))
				rep.Item("Path", rulePath)
				fmt.Println(string(data))
				return nil
			}
		}
	}

	// Check workflow packs
	packDir := filepath.Join(config.AgentDir, "workflow-packs", target)
	if _, err := os.Stat(packDir); err == nil {
		rep.Item("Type", "workflow pack")
		rep.Item("Path", packDir)
		readme := filepath.Join(packDir, "README.md")
		if data, err := os.ReadFile(readme); err == nil {
			fmt.Println(string(data))
		}
		return nil
	}

	rep.Item("Status", fmt.Sprintf("'%s' not found in toolkit, rules, or workflow packs", target))
	rep.Item("Hint", "Try: rice-rail explain constitution.yaml")

	return nil
}
