package adapters

// RuleSpec is the full metadata for a project rule (PRD Section 15).
type RuleSpec struct {
	ID              string   `yaml:"id" json:"id"`
	Title           string   `yaml:"title" json:"title"`
	Description     string   `yaml:"description" json:"description"`
	Rationale       string   `yaml:"rationale" json:"rationale"`
	Severity        string   `yaml:"severity" json:"severity"`     // BLOCKING, WARNING, INFO, DISABLED
	Category        string   `yaml:"category" json:"category"`     // formatting, naming, imports, layering, test_discipline, code_smell, ddd, security, advisory
	FixKind         string   `yaml:"fix_kind" json:"fix_kind"`     // NONE, SAFE_AUTOFIX, UNSAFE_AUTOFIX, CODEMOD, AI_REPAIR, HUMAN_REVIEW
	EvidenceSource  string   `yaml:"evidence_source" json:"evidence_source"`
	Languages       []string `yaml:"applies_to_languages" json:"applies_to_languages"`
	Paths           []string `yaml:"applies_to_paths,omitempty" json:"applies_to_paths,omitempty"`
	Origin          string   `yaml:"origin" json:"origin"` // inferred, user_stated, company_pack, generated
	WaiverPolicy    string   `yaml:"waiver_policy" json:"waiver_policy"`
}

// CodemodSpec is the full metadata for a project codemod (PRD Section 16).
type CodemodSpec struct {
	ID             string   `yaml:"id" json:"id"`
	Purpose        string   `yaml:"purpose" json:"purpose"`
	Languages      []string `yaml:"target_languages" json:"target_languages"`
	Preconditions  []string `yaml:"preconditions" json:"preconditions"`
	Postconditions []string `yaml:"postconditions" json:"postconditions"`
	SafetyClass    string   `yaml:"safety_class" json:"safety_class"` // SAFE, REVIEW_REQUIRED, UNSAFE
	Rollback       string   `yaml:"rollback_guidance" json:"rollback_guidance"`
	TestRequired   bool     `yaml:"test_requirements" json:"test_requirements"`
	Engine         string   `yaml:"engine" json:"engine"` // ast-grep, comby, openrewrite, jscodeshift
}

// RuleCatalog is the collection of all rules in a project.
type RuleCatalog struct {
	Rules []RuleSpec `yaml:"rules" json:"rules"`
}

// CodemodCatalog is the collection of all codemods in a project.
type CodemodCatalog struct {
	Codemods []CodemodSpec `yaml:"codemods" json:"codemods"`
}
