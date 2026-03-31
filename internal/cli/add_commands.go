package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/reporting"
)

var addSkillCmd = &cobra.Command{
	Use:   "add-skill <pack-name>",
	Short: "Add a workflow pack to the agent directory",
	Long:  "Copy or link a workflow pack into .agent/workflow-packs/<name>/ and create a README.md stub.",
	Args:  cobra.ExactArgs(1),
	RunE:  runAddSkill,
}

var addMCPCmd = &cobra.Command{
	Use:   "add-mcp <server-name>",
	Short: "Add an MCP server to the constitution",
	Long:  "Add a server name to the constitution's mcp.optional_servers list and save the updated constitution.",
	Args:  cobra.ExactArgs(1),
	RunE:  runAddMCP,
}

func runAddSkill(cmd *cobra.Command, args []string) error {
	packName := args[0]
	rep := reporting.New(getFormat())

	rep.Section("Adding Workflow Pack")

	packDir := filepath.Join(config.AgentDir, "workflow-packs", packName)
	if err := os.MkdirAll(packDir, 0755); err != nil {
		return fmt.Errorf("creating workflow pack directory %s: %w", packDir, err)
	}
	rep.Status(packDir, "CREATED")

	readmePath := filepath.Join(packDir, "README.md")
	readmeContent := fmt.Sprintf(`# Workflow Pack: %s

## Purpose

Agent guidance for the %s workflow pack.

## Instructions

Follow the project constitution in .project-toolkit/constitution.yaml.
`, packName, packName)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("writing README: %w", err)
	}
	rep.Status(readmePath, "CREATED")

	rep.Section("Add Skill Complete")
	rep.Item("Pack", packName)
	rep.Item("Location", packDir)

	return nil
}

func runAddMCP(cmd *cobra.Command, args []string) error {
	serverName := args[0]
	p := paths()
	rep := reporting.New(getFormat())

	rep.Section("Adding MCP Server")

	// Load current constitution
	c, err := constitution.Load(p.Constitution)
	if err != nil {
		return fmt.Errorf("loading constitution (run 'rice-rail init' first): %w", err)
	}

	// Check for duplicates
	for _, s := range c.MCP.OptionalServers {
		if s == serverName {
			rep.Status(serverName, "ALREADY EXISTS")
			return nil
		}
	}

	// Add server
	c.MCP.OptionalServers = append(c.MCP.OptionalServers, serverName)

	// Save updated constitution
	if err := constitution.Save(c, p.Constitution); err != nil {
		return fmt.Errorf("saving constitution: %w", err)
	}

	rep.Status(serverName, "ADDED")
	rep.Status(p.Constitution, "SAVED")

	rep.Section("Add MCP Complete")
	rep.Item("Server", serverName)
	rep.Item("Total optional servers", fmt.Sprintf("%d", len(c.MCP.OptionalServers)))

	return nil
}
