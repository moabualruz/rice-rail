package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mkh/rice-railing/internal/config"
	"github.com/mkh/rice-railing/internal/constitution"
	"github.com/mkh/rice-railing/internal/exec"
	"github.com/mkh/rice-railing/internal/reporting"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose broken toolkit install or toolchain drift",
	Long:  "Check toolkit directory, constitution, tools, wrappers, workflow packs, and version consistency. Report each check as PASS/FAIL/WARN.",
	RunE:  runDoctor,
}

// doctorCheck records one diagnostic check.
type doctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // PASS, FAIL, WARN
	Detail string `json:"detail,omitempty"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	p := paths()
	rep := reporting.New(getFormat())
	rep.Section("Doctor: Toolkit Health Check")

	var checks []doctorCheck
	hasCriticalFail := false

	// 1. Check .project-toolkit/ directory exists
	checks = append(checks, checkDirExists(p.ToolkitDir, "toolkit directory"))

	// 2. Check constitution exists and parses
	constitutionCheck, c := checkConstitutionParseable(p.Constitution)
	checks = append(checks, constitutionCheck)
	if constitutionCheck.Status == "FAIL" {
		hasCriticalFail = true
	}

	// 3. Check preferred tools on PATH
	if c != nil {
		checks = append(checks, checkToolsOnPath(c)...)
	}

	// 4. Check wrapper scripts in bin/
	checks = append(checks, checkWrapperScripts()...)

	// 5. Check workflow packs
	checks = append(checks, checkWorkflowPacks()...)

	// 6. Check toolkit version consistency
	checks = append(checks, checkToolkitVersion(p))

	// Print results
	for _, chk := range checks {
		rep.Status(chk.Name, chk.Status)
		if chk.Detail != "" && chk.Status != "PASS" {
			rep.Item("  detail", chk.Detail)
		}
		if chk.Status == "FAIL" {
			hasCriticalFail = true
		}
	}

	rep.Section("Summary")
	pass, fail, warn := 0, 0, 0
	for _, chk := range checks {
		switch chk.Status {
		case "PASS":
			pass++
		case "FAIL":
			fail++
		case "WARN":
			warn++
		}
	}
	rep.Item("Total checks", fmt.Sprintf("%d", len(checks)))
	rep.Item("PASS", fmt.Sprintf("%d", pass))
	rep.Item("FAIL", fmt.Sprintf("%d", fail))
	rep.Item("WARN", fmt.Sprintf("%d", warn))

	if hasCriticalFail {
		return fmt.Errorf("doctor found %d critical failure(s); run 'rice-rail init' or 'rice-rail regenerate' to fix", fail)
	}
	return nil
}

func checkDirExists(path, label string) doctorCheck {
	info, err := os.Stat(path)
	if err != nil {
		return doctorCheck{Name: label, Status: "FAIL", Detail: path + " not found"}
	}
	if !info.IsDir() {
		return doctorCheck{Name: label, Status: "FAIL", Detail: path + " exists but is not a directory"}
	}
	return doctorCheck{Name: label, Status: "PASS"}
}

func checkConstitutionParseable(path string) (doctorCheck, *constitution.Constitution) {
	c, err := constitution.Load(path)
	if err != nil {
		return doctorCheck{Name: "constitution", Status: "FAIL", Detail: err.Error()}, nil
	}
	return doctorCheck{Name: "constitution", Status: "PASS"}, c
}

func checkToolsOnPath(c *constitution.Constitution) []doctorCheck {
	var checks []doctorCheck
	allTools := collectPreferredTools(c)

	for _, tool := range allTools {
		_, found := exec.Which(tool)
		status := "PASS"
		detail := ""
		if !found {
			status = "WARN"
			detail = tool + " not found on PATH"
		}
		checks = append(checks, doctorCheck{
			Name:   "tool: " + tool,
			Status: status,
			Detail: detail,
		})
	}
	return checks
}

func collectPreferredTools(c *constitution.Constitution) []string {
	seen := map[string]bool{}
	var tools []string
	addUnique := func(names []string) {
		for _, n := range names {
			if !seen[n] {
				seen[n] = true
				tools = append(tools, n)
			}
		}
	}
	addUnique(c.Tools.Linters)
	addUnique(c.Tools.Formatters)
	addUnique(c.Tools.Typecheckers)
	addUnique(c.Tools.TestRunners)
	addUnique(c.Tools.CodemodEngines)
	addUnique(c.Tools.RuleEngines)
	return tools
}

func checkWrapperScripts() []doctorCheck {
	var checks []doctorCheck
	wrappers := []string{
		"bin/rice-rail-check",
		"bin/rice-rail-fix",
		"bin/rice-rail-baseline",
		"bin/rice-rail-cycle",
		"bin/rice-rail-report",
		"bin/rice-rail-explain",
	}

	for _, w := range wrappers {
		info, err := os.Stat(w)
		if err != nil {
			checks = append(checks, doctorCheck{
				Name:   "wrapper: " + w,
				Status: "WARN",
				Detail: "not found (run 'rice-rail build-toolkit' to generate)",
			})
			continue
		}
		if info.Mode()&0111 == 0 {
			checks = append(checks, doctorCheck{
				Name:   "wrapper: " + w,
				Status: "WARN",
				Detail: "exists but not executable",
			})
			continue
		}
		checks = append(checks, doctorCheck{Name: "wrapper: " + w, Status: "PASS"})
	}
	return checks
}

func checkWorkflowPacks() []doctorCheck {
	var checks []doctorCheck
	packsDir := filepath.Join(config.AgentDir, "workflow-packs")

	info, err := os.Stat(packsDir)
	if err != nil || !info.IsDir() {
		return []doctorCheck{{
			Name:   "workflow packs directory",
			Status: "WARN",
			Detail: packsDir + " not found",
		}}
	}

	entries, err := os.ReadDir(packsDir)
	if err != nil {
		return []doctorCheck{{
			Name:   "workflow packs directory",
			Status: "WARN",
			Detail: "cannot read " + packsDir,
		}}
	}

	if len(entries) == 0 {
		return []doctorCheck{{
			Name:   "workflow packs",
			Status: "WARN",
			Detail: "no workflow packs found in " + packsDir,
		}}
	}

	for _, e := range entries {
		if e.IsDir() {
			checks = append(checks, doctorCheck{
				Name:   "workflow pack: " + e.Name(),
				Status: "PASS",
			})
		}
	}
	return checks
}

// toolkitVersionInfo represents the stored toolkit version state.
type toolkitVersionInfo struct {
	ToolkitVersion      string `json:"toolkit_version"`
	ConstitutionVersion int    `json:"constitution_version"`
}

func checkToolkitVersion(p config.Paths) doctorCheck {
	versionFile := filepath.Join(config.StateDir, "toolkit-version.json")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return doctorCheck{
			Name:   "toolkit version",
			Status: "WARN",
			Detail: "no version file at " + versionFile + " (not tracked yet)",
		}
	}

	var ver toolkitVersionInfo
	if err := json.Unmarshal(data, &ver); err != nil {
		return doctorCheck{
			Name:   "toolkit version",
			Status: "WARN",
			Detail: "cannot parse " + versionFile,
		}
	}

	c, err := constitution.Load(p.Constitution)
	if err != nil {
		return doctorCheck{
			Name:   "toolkit version",
			Status: "WARN",
			Detail: "cannot load constitution to compare versions",
		}
	}

	if c.Version != ver.ConstitutionVersion {
		return doctorCheck{
			Name:   "toolkit version",
			Status: "WARN",
			Detail: fmt.Sprintf("constitution version (%d) differs from toolkit version (%d); run 'rice-rail regenerate'", c.Version, ver.ConstitutionVersion),
		}
	}

	return doctorCheck{Name: "toolkit version", Status: "PASS"}
}
