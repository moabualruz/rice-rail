package constitution

// Constitution is the versioned source of truth for a project's engineering doctrine.
type Constitution struct {
	Version      int              `yaml:"version" json:"version"`
	Project      ProjectInfo      `yaml:"project" json:"project"`
	Architecture ArchitectureSpec `yaml:"architecture" json:"architecture"`
	Quality      QualitySpec      `yaml:"quality" json:"quality"`
	Automation   AutomationSpec   `yaml:"automation" json:"automation"`
	Tools        ToolPreferences  `yaml:"tool_preferences" json:"tool_preferences"`
	Workflow     WorkflowSpec     `yaml:"workflow" json:"workflow"`
	MCP          MCPSpec          `yaml:"mcp" json:"mcp"`
	Agent        AgentSpec        `yaml:"agent" json:"agent"`
}

type ProjectInfo struct {
	Name            string   `yaml:"name" json:"name"`
	RepoType        string   `yaml:"repo_type" json:"repo_type"` // single, monorepo, hybrid
	Languages       []string `yaml:"languages" json:"languages"`
	PackageManagers []string `yaml:"package_managers" json:"package_managers"`
	BuildSystems    []string `yaml:"build_systems" json:"build_systems"`
	RuntimeTargets  []string `yaml:"runtime_targets" json:"runtime_targets"`
	PrimaryOS       []string `yaml:"primary_os" json:"primary_os"`
}

type ArchitectureSpec struct {
	TargetStyle      string           `yaml:"target_style" json:"target_style"`
	BoundedContexts  []string         `yaml:"bounded_contexts" json:"bounded_contexts"`
	Modules          []string         `yaml:"modules" json:"modules"`
	Layering         LayeringSpec     `yaml:"layering" json:"layering"`
	DependencyPolicy DependencyPolicy `yaml:"dependency_policy" json:"dependency_policy"`
	DomainRules      DomainRules      `yaml:"domain_rules" json:"domain_rules"`
}

type LayeringSpec struct {
	Enabled               bool                  `yaml:"enabled" json:"enabled"`
	Layers                []string              `yaml:"layers" json:"layers"`
	ForbiddenDependencies []ForbiddenDependency `yaml:"forbidden_dependencies" json:"forbidden_dependencies"`
}

type ForbiddenDependency struct {
	From string `yaml:"from" json:"from"`
	To   string `yaml:"to" json:"to"`
}

type DependencyPolicy struct {
	NoCycles                    bool `yaml:"no_cycles" json:"no_cycles"`
	RestrictCrossContextImports bool `yaml:"restrict_cross_context_imports" json:"restrict_cross_context_imports"`
	AllowSharedKernel           bool `yaml:"allow_shared_kernel" json:"allow_shared_kernel"`
}

type DomainRules struct {
	NoFrameworkTypesInDomain     bool `yaml:"no_framework_types_in_domain" json:"no_framework_types_in_domain"`
	RepositoriesAsInterfacesOnly bool `yaml:"repositories_as_interfaces_only" json:"repositories_as_interfaces_only"`
	EntitiesImmutablePreference  bool `yaml:"entities_immutable_preference" json:"entities_immutable_preference"`
	AdaptersAtEdgesOnly          bool `yaml:"adapters_at_edges_only" json:"adapters_at_edges_only"`
}

type QualitySpec struct {
	SafetyMode              string   `yaml:"safety_mode" json:"safety_mode"` // strict, balanced, aggressive
	BlockOn                 []string `yaml:"block_on" json:"block_on"`
	AdvisoryOn              []string `yaml:"advisory_on" json:"advisory_on"`
	RequireChangedCodeTests bool     `yaml:"require_changed_code_tests" json:"require_changed_code_tests"`
	MaxChangedFilesPerCycle int      `yaml:"max_changed_files_per_cycle" json:"max_changed_files_per_cycle"`
	MaxChangedLinesPerCycle int      `yaml:"max_changed_lines_per_cycle" json:"max_changed_lines_per_cycle"`
}

type AutomationSpec struct {
	AllowSafeAutofix                  bool   `yaml:"allow_safe_autofix" json:"allow_safe_autofix"`
	AllowUnsafeAutofix                bool   `yaml:"allow_unsafe_autofix" json:"allow_unsafe_autofix"`
	AllowGeneratedCodemods            bool   `yaml:"allow_generated_codemods" json:"allow_generated_codemods"`
	AllowCrossModuleRewrites          bool   `yaml:"allow_cross_module_rewrites" json:"allow_cross_module_rewrites"`
	AllowRuleSuppressionWithoutWaiver bool   `yaml:"allow_rule_suppression_without_waiver" json:"allow_rule_suppression_without_waiver"`
	BaselineModeDefault               string `yaml:"baseline_mode_default" json:"baseline_mode_default"`
	FeatureModeDefault                string `yaml:"feature_mode_default" json:"feature_mode_default"`
}

type ToolPreferences struct {
	Linters        []string     `yaml:"preferred_linters" json:"preferred_linters"`
	Formatters     []string     `yaml:"preferred_formatters" json:"preferred_formatters"`
	Typecheckers   []string     `yaml:"preferred_typecheckers" json:"preferred_typecheckers"`
	TestRunners    []string     `yaml:"preferred_test_runners" json:"preferred_test_runners"`
	CodemodEngines []string     `yaml:"preferred_codemod_engines" json:"preferred_codemod_engines"`
	RuleEngines    []string     `yaml:"preferred_rule_engines" json:"preferred_rule_engines"`
	Custom         []CustomTool `yaml:"custom,omitempty" json:"custom,omitempty"`
}

// CustomTool defines a user-provided tool that rice-rail should integrate.
// This allows any CLI tool to participate in check/fix/test flows without
// writing a Go adapter.
type CustomTool struct {
	Name       string   `yaml:"name" json:"name"`
	Binary     string   `yaml:"binary" json:"binary"`                             // command name or path
	Role       string   `yaml:"role" json:"role"`                                 // linter, formatter, typechecker, test_runner, rule_engine, codemod
	Languages  []string `yaml:"languages,omitempty" json:"languages,omitempty"`   // empty = all
	CheckCmd   []string `yaml:"check_cmd,omitempty" json:"check_cmd,omitempty"`   // args for check mode (read-only)
	FixCmd     []string `yaml:"fix_cmd,omitempty" json:"fix_cmd,omitempty"`       // args for fix mode (writes files)
	TestCmd    []string `yaml:"test_cmd,omitempty" json:"test_cmd,omitempty"`     // args for test mode
	OutputFmt  string   `yaml:"output_format,omitempty" json:"output_format,omitempty"` // json, text, sarif (default: text)
	SuccessExit int     `yaml:"success_exit,omitempty" json:"success_exit,omitempty"`   // exit code for success (default: 0)
	SafetyClass string  `yaml:"safety_class,omitempty" json:"safety_class,omitempty"`   // safe, review_required, unsafe (default: safe for fix, n/a for check)
}

type WorkflowSpec struct {
	RequirePlanBeforeCycle                bool `yaml:"require_plan_before_cycle" json:"require_plan_before_cycle"`
	SeparateBaselineAndFeatureWork        bool `yaml:"separate_baseline_and_feature_work" json:"separate_baseline_and_feature_work"`
	RequireHumanAckForConstitutionChanges bool `yaml:"require_human_ack_for_constitution_changes" json:"require_human_ack_for_constitution_changes"`
	GenerateCIIntegration                 bool `yaml:"generate_ci_integration" json:"generate_ci_integration"`
	GenerateLocalWrappers                 bool `yaml:"generate_local_wrappers" json:"generate_local_wrappers"`
}

type MCPSpec struct {
	Enabled         bool     `yaml:"enabled" json:"enabled"`
	OptionalServers []string `yaml:"optional_servers" json:"optional_servers"`
	RequiredServers []string `yaml:"required_servers" json:"required_servers"`
	AccessPolicy    string   `yaml:"access_policy" json:"access_policy"` // local_only, allow_remote_optional
}

type AgentSpec struct {
	DefaultAdapter        string                   `yaml:"default_adapter" json:"default_adapter"`
	EnabledAdapters       []string                 `yaml:"enabled_adapters" json:"enabled_adapters"`
	WorkflowPackRoot      string                   `yaml:"workflow_pack_root" json:"workflow_pack_root"`
	AllowRemoteConnectors bool                     `yaml:"allow_remote_connectors" json:"allow_remote_connectors"`
	Adapters              map[string]AdapterConfig `yaml:"adapters" json:"adapters"`
}

type AdapterConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}
