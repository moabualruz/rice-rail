package profiling

// RepoProfile is the inferred profile of a repository.
type RepoProfile struct {
	Languages       []DetectedItem `yaml:"languages" json:"languages"`
	PackageManagers []DetectedItem `yaml:"package_managers" json:"package_managers"`
	BuildSystems    []DetectedItem `yaml:"build_systems" json:"build_systems"`
	RepoTopology    string         `yaml:"repo_topology" json:"repo_topology"` // single, monorepo, hybrid
	Frameworks      []DetectedItem `yaml:"frameworks" json:"frameworks"`
	Tooling         ToolingProfile `yaml:"tooling" json:"tooling"`
	CI              []DetectedItem `yaml:"ci" json:"ci"`
	ArchHints       []ArchHint     `yaml:"arch_hints" json:"arch_hints"`
	Evidence        []Evidence     `yaml:"evidence" json:"evidence"`
	Unresolved      []string       `yaml:"unresolved" json:"unresolved"`
}

// DetectedItem is something found with a confidence score.
type DetectedItem struct {
	Name       string  `yaml:"name" json:"name"`
	Confidence float64 `yaml:"confidence" json:"confidence"` // 0.0-1.0
	Evidence   string  `yaml:"evidence" json:"evidence"`
}

// ToolingProfile groups detected dev tools by category.
type ToolingProfile struct {
	Linters      []DetectedTool `yaml:"linters" json:"linters"`
	Formatters   []DetectedTool `yaml:"formatters" json:"formatters"`
	Typecheckers []DetectedTool `yaml:"typecheckers" json:"typecheckers"`
	TestRunners  []DetectedTool `yaml:"test_runners" json:"test_runners"`
	RuleEngines  []DetectedTool `yaml:"rule_engines" json:"rule_engines"`
	Codemods     []DetectedTool `yaml:"codemods" json:"codemods"`
	Security     []DetectedTool `yaml:"security" json:"security"`
}

// DetectedTool is a development tool found in the repo or system.
type DetectedTool struct {
	Name       string  `yaml:"name" json:"name"`
	Category   string  `yaml:"category" json:"category"`
	ConfigFile string  `yaml:"config_file,omitempty" json:"config_file,omitempty"`
	Installed  bool    `yaml:"installed" json:"installed"`
	Confidence float64 `yaml:"confidence" json:"confidence"`
	Evidence   string  `yaml:"evidence" json:"evidence"`
}

// ArchHint is an architecture inference from folder structure or imports.
type ArchHint struct {
	Pattern    string  `yaml:"pattern" json:"pattern"`
	Suggests   string  `yaml:"suggests" json:"suggests"`
	Confidence float64 `yaml:"confidence" json:"confidence"`
	Evidence   string  `yaml:"evidence" json:"evidence"`
}

// Evidence records what was found and where.
type Evidence struct {
	File    string `yaml:"file" json:"file"`
	Kind    string `yaml:"kind" json:"kind"` // config, manifest, structure, content
	Finding string `yaml:"finding" json:"finding"`
}
