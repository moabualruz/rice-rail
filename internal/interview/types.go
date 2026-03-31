package interview

// Question represents a single adaptive interview question.
type Question struct {
	ID           string   `yaml:"id" json:"id"`
	Group        string   `yaml:"group" json:"group"`
	Text         string   `yaml:"text" json:"text"`
	HelpText     string   `yaml:"help_text,omitempty" json:"help_text,omitempty"`
	Type         string   `yaml:"type" json:"type"` // choice, multi_choice, confirm, text, number
	Options      []Option `yaml:"options,omitempty" json:"options,omitempty"`
	Default      string   `yaml:"default,omitempty" json:"default,omitempty"`
	InferredFrom string   `yaml:"inferred_from,omitempty" json:"inferred_from,omitempty"`
	Confidence   float64  `yaml:"confidence,omitempty" json:"confidence,omitempty"`
	SkipWhen     []string `yaml:"skip_when,omitempty" json:"skip_when,omitempty"`       // conditions that skip this question
	RequireWhen  []string `yaml:"require_when,omitempty" json:"require_when,omitempty"` // conditions that require this question
}

// Option is a selectable answer for choice questions.
type Option struct {
	Value       string `yaml:"value" json:"value"`
	Label       string `yaml:"label" json:"label"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Answer records a user's response to a question.
type Answer struct {
	QuestionID string   `yaml:"question_id" json:"question_id"`
	Value      string   `yaml:"value" json:"value"`
	Values     []string `yaml:"values,omitempty" json:"values,omitempty"` // for multi_choice
	Accepted   bool     `yaml:"accepted" json:"accepted"`                 // true if user accepted inferred default
	Source     string   `yaml:"source" json:"source"`                     // user, inferred, seed, preset
}

// Transcript is the full interview record.
type Transcript struct {
	Mode    string   `yaml:"mode" json:"mode"` // quick, normal, strict
	Answers []Answer `yaml:"answers" json:"answers"`
}

// Seed is a preset file that pre-answers questions.
type Seed struct {
	Answers map[string]string   `yaml:"answers" json:"answers"`
	Multi   map[string][]string `yaml:"multi,omitempty" json:"multi,omitempty"`
}
