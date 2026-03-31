package interview

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mkh/rice-railing/internal/profiling"
)

// Mode controls how many questions are asked.
type Mode int

const (
	ModeNormal Mode = iota
	ModeQuick       // accept high-confidence inferred answers, minimal questions
	ModeStrict      // ask all question groups
)

// Engine runs the adaptive interview.
type Engine struct {
	Mode       Mode
	Profile    *profiling.RepoProfile
	Seed       *Seed
	Answers    map[string]*Answer
	Transcript Transcript
}

// NewEngine creates an interview engine with optional profile and seed.
func NewEngine(mode Mode, profile *profiling.RepoProfile, seed *Seed) *Engine {
	return &Engine{
		Mode:    mode,
		Profile: profile,
		Seed:    seed,
		Answers: make(map[string]*Answer),
	}
}

// Prompter is the interface for asking questions interactively.
type Prompter interface {
	AskChoice(question string, options []Option, defaultVal string) (string, error)
	AskMultiChoice(question string, options []Option, defaults []string) ([]string, error)
	AskConfirm(question string, defaultVal bool) (bool, error)
	AskText(question string, defaultVal string) (string, error)
	AskNumber(question string, defaultVal int) (int, error)
	ShowInferred(question string, value string) (bool, error) // returns true if accepted
}

// Run executes the interview and returns the transcript.
func (e *Engine) Run(prompter Prompter) (*Transcript, error) {
	e.applyInferences()
	e.applySeed()

	questions := QuestionCatalog()
	modeStr := "normal"
	if e.Mode == ModeQuick {
		modeStr = "quick"
	} else if e.Mode == ModeStrict {
		modeStr = "strict"
	}
	e.Transcript.Mode = modeStr

	for _, q := range questions {
		if e.shouldSkip(q) {
			continue
		}

		// If already answered by seed or inference
		if existing, ok := e.Answers[q.ID]; ok {
			if e.Mode == ModeQuick && existing.Confidence() >= 0.8 {
				e.Transcript.Answers = append(e.Transcript.Answers, *existing)
				continue
			}
			// Show inferred value and let user accept/edit
			accepted, err := prompter.ShowInferred(q.Text, existing.Value)
			if err != nil {
				return nil, fmt.Errorf("question %s: %w", q.ID, err)
			}
			if accepted {
				existing.Accepted = true
				e.Transcript.Answers = append(e.Transcript.Answers, *existing)
				continue
			}
		}

		answer, err := e.ask(prompter, q)
		if err != nil {
			return nil, fmt.Errorf("question %s: %w", q.ID, err)
		}

		e.Answers[q.ID] = answer
		e.Transcript.Answers = append(e.Transcript.Answers, *answer)
	}

	return &e.Transcript, nil
}

// Confidence returns the confidence of an answer based on its source.
func (a *Answer) Confidence() float64 {
	switch a.Source {
	case "inferred":
		return 0.8
	case "seed":
		return 0.95
	case "preset":
		return 0.9
	case "user":
		return 1.0
	default:
		return 0.5
	}
}

func (e *Engine) applyInferences() {
	if e.Profile == nil {
		return
	}

	// Infer repo topology
	if e.Profile.RepoTopology != "" {
		purpose := "application"
		if e.Profile.RepoTopology == "monorepo" {
			purpose = "monorepo"
		}
		e.Answers["repo_purpose"] = &Answer{
			QuestionID: "repo_purpose",
			Value:      purpose,
			Source:     "inferred",
		}
	}

	// Infer architecture from folder hints
	for _, hint := range e.Profile.ArchHints {
		switch {
		case strings.Contains(hint.Suggests, "hexagonal") || strings.Contains(hint.Suggests, "ports-and-adapters"):
			e.Answers["arch_target"] = &Answer{QuestionID: "arch_target", Value: "hexagonal", Source: "inferred"}
		case strings.Contains(hint.Suggests, "clean architecture"):
			e.Answers["arch_target"] = &Answer{QuestionID: "arch_target", Value: "clean", Source: "inferred"}
		case strings.Contains(hint.Suggests, "domain-driven"):
			e.Answers["arch_target"] = &Answer{QuestionID: "arch_target", Value: "modular_monolith", Source: "inferred"}
		}
	}
}

func (e *Engine) applySeed() {
	if e.Seed == nil {
		return
	}
	for qID, val := range e.Seed.Answers {
		e.Answers[qID] = &Answer{
			QuestionID: qID,
			Value:      val,
			Source:     "seed",
		}
	}
	for qID, vals := range e.Seed.Multi {
		e.Answers[qID] = &Answer{
			QuestionID: qID,
			Values:     vals,
			Value:      strings.Join(vals, ","),
			Source:     "seed",
		}
	}
}

func (e *Engine) shouldSkip(q Question) bool {
	// In strict mode, never skip
	if e.Mode == ModeStrict {
		return false
	}

	// Check skip conditions
	for _, cond := range q.SkipWhen {
		if e.conditionMet(cond) {
			return true
		}
	}

	// Check require conditions — if any exist, at least one must be met
	if len(q.RequireWhen) > 0 {
		met := false
		for _, cond := range q.RequireWhen {
			if e.conditionMet(cond) {
				met = true
				break
			}
		}
		if !met {
			return true
		}
	}

	return false
}

func (e *Engine) conditionMet(cond string) bool {
	parts := strings.SplitN(cond, "=", 2)
	if len(parts) != 2 {
		return false
	}
	answer, ok := e.Answers[parts[0]]
	if !ok {
		return false
	}
	return answer.Value == parts[1]
}

func (e *Engine) ask(prompter Prompter, q Question) (*Answer, error) {
	answer := &Answer{
		QuestionID: q.ID,
		Source:     "user",
	}

	switch q.Type {
	case "choice":
		val, err := prompter.AskChoice(q.Text, q.Options, q.Default)
		if err != nil {
			return nil, err
		}
		answer.Value = val

	case "multi_choice":
		defaults := strings.Split(q.Default, ",")
		vals, err := prompter.AskMultiChoice(q.Text, q.Options, defaults)
		if err != nil {
			return nil, err
		}
		answer.Values = vals
		answer.Value = strings.Join(vals, ",")

	case "confirm":
		def := q.Default == "true"
		val, err := prompter.AskConfirm(q.Text, def)
		if err != nil {
			return nil, err
		}
		answer.Value = strconv.FormatBool(val)

	case "text":
		val, err := prompter.AskText(q.Text, q.Default)
		if err != nil {
			return nil, err
		}
		answer.Value = val

	case "number":
		def, _ := strconv.Atoi(q.Default)
		val, err := prompter.AskNumber(q.Text, def)
		if err != nil {
			return nil, err
		}
		answer.Value = strconv.Itoa(val)
	}

	return answer, nil
}
