package discovery

// ToolInventory is the full catalog of discovered tools.
type ToolInventory struct {
	Tools []InventoryEntry `yaml:"tools" json:"tools"`
}

// InventoryEntry describes a single discovered tool.
type InventoryEntry struct {
	Name               string   `yaml:"name" json:"name"`
	Category           string   `yaml:"category" json:"category"` // linter, formatter, typechecker, test_runner, rule_engine, codemod, security, build, ci
	AdapterClass       string   `yaml:"adapter_class" json:"adapter_class"`
	SupportedLanguages []string `yaml:"supported_languages" json:"supported_languages"`
	InvocationMethod   string   `yaml:"invocation_method" json:"invocation_method"` // cli, config, library
	InstalledLocally   bool     `yaml:"installed_locally" json:"installed_locally"`
	ConfiguredInRepo   bool     `yaml:"configured_in_repo" json:"configured_in_repo"`
	Role               string   `yaml:"role" json:"role"`
	SafetyClass        string   `yaml:"safety_class" json:"safety_class"` // safe, review_required, unsafe
	Status             string   `yaml:"status" json:"status"`             // available, configured, partially_configured, missing, optional
	Evidence           string   `yaml:"evidence" json:"evidence"`
	RecommendedAction  string   `yaml:"recommended_action" json:"recommended_action"`
}

// CapabilityMatrix maps desired capabilities to their coverage status.
type CapabilityMatrix struct {
	Capabilities []CapabilityEntry `yaml:"capabilities" json:"capabilities"`
}

// CapabilityEntry describes a single desired capability and its status.
type CapabilityEntry struct {
	Name      string `yaml:"name" json:"name"`
	Category  string `yaml:"category" json:"category"`
	Status    string `yaml:"status" json:"status"` // PRESENT_READY, PRESENT_NEEDS_CONFIG, PRESENT_NEEDS_WRAPPER, NEEDS_PROJECT_RULE, NEEDS_PROJECT_CODEMOD, NEEDS_NEW_TOOL, ADVISORY_ONLY, DEFERRED
	Tool      string `yaml:"tool,omitempty" json:"tool,omitempty"`
	Reasoning string `yaml:"reasoning" json:"reasoning"`
	Priority  int    `yaml:"priority" json:"priority"` // 1=critical, 2=important, 3=nice-to-have
	Risk      string `yaml:"risk" json:"risk"`         // low, medium, high
	Strategy  string `yaml:"strategy" json:"strategy"`
}
