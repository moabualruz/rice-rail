package interview

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

// TerminalPrompter implements Prompter using Charm's huh library.
type TerminalPrompter struct{}

func (t *TerminalPrompter) AskChoice(question string, options []Option, defaultVal string) (string, error) {
	var result string
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		label := o.Label
		if o.Description != "" {
			label = fmt.Sprintf("%s — %s", o.Label, o.Description)
		}
		opts[i] = huh.NewOption(label, o.Value)
	}

	err := huh.NewSelect[string]().
		Title(question).
		Options(opts...).
		Value(&result).
		Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (t *TerminalPrompter) AskMultiChoice(question string, options []Option, defaults []string) ([]string, error) {
	var result []string
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o.Label, o.Value)
	}

	err := huh.NewMultiSelect[string]().
		Title(question).
		Options(opts...).
		Value(&result).
		Run()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *TerminalPrompter) AskConfirm(question string, defaultVal bool) (bool, error) {
	var result bool
	err := huh.NewConfirm().
		Title(question).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	if err != nil {
		return false, err
	}
	return result, nil
}

func (t *TerminalPrompter) AskText(question string, defaultVal string) (string, error) {
	var result string
	input := huh.NewInput().
		Title(question).
		Value(&result)
	if defaultVal != "" {
		result = defaultVal
	}
	if err := input.Run(); err != nil {
		return "", err
	}
	return result, nil
}

func (t *TerminalPrompter) AskNumber(question string, defaultVal int) (int, error) {
	var result string
	result = strconv.Itoa(defaultVal)
	err := huh.NewInput().
		Title(question).
		Value(&result).
		Validate(func(s string) error {
			_, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("enter a number")
			}
			return nil
		}).
		Run()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(result)
}

func (t *TerminalPrompter) ShowInferred(question string, value string) (bool, error) {
	var accept bool
	err := huh.NewConfirm().
		Title(fmt.Sprintf("%s [inferred: %s]", question, value)).
		Affirmative("Accept").
		Negative("Change").
		Value(&accept).
		Run()
	if err != nil {
		return false, err
	}
	return accept, nil
}

// NonInteractivePrompter always accepts defaults. Used for CI and seed mode.
type NonInteractivePrompter struct {
	Defaults map[string]string
}

func (n *NonInteractivePrompter) AskChoice(question string, options []Option, defaultVal string) (string, error) {
	return defaultVal, nil
}

func (n *NonInteractivePrompter) AskMultiChoice(question string, options []Option, defaults []string) ([]string, error) {
	return defaults, nil
}

func (n *NonInteractivePrompter) AskConfirm(question string, defaultVal bool) (bool, error) {
	return defaultVal, nil
}

func (n *NonInteractivePrompter) AskText(question string, defaultVal string) (string, error) {
	return defaultVal, nil
}

func (n *NonInteractivePrompter) AskNumber(question string, defaultVal int) (int, error) {
	return defaultVal, nil
}

func (n *NonInteractivePrompter) ShowInferred(question string, value string) (bool, error) {
	return true, nil
}

// RecordingPrompter wraps another prompter and records all Q&A for transcript generation.
type RecordingPrompter struct {
	Inner   Prompter
	Records []QARecord
}

// QARecord is a single question-answer pair.
type QARecord struct {
	Question string
	Answer   string
}

func (r *RecordingPrompter) AskChoice(question string, options []Option, defaultVal string) (string, error) {
	val, err := r.Inner.AskChoice(question, options, defaultVal)
	if err == nil {
		r.Records = append(r.Records, QARecord{Question: question, Answer: val})
	}
	return val, err
}

func (r *RecordingPrompter) AskMultiChoice(question string, options []Option, defaults []string) ([]string, error) {
	vals, err := r.Inner.AskMultiChoice(question, options, defaults)
	if err == nil {
		r.Records = append(r.Records, QARecord{Question: question, Answer: strings.Join(vals, ", ")})
	}
	return vals, err
}

func (r *RecordingPrompter) AskConfirm(question string, defaultVal bool) (bool, error) {
	val, err := r.Inner.AskConfirm(question, defaultVal)
	if err == nil {
		r.Records = append(r.Records, QARecord{Question: question, Answer: strconv.FormatBool(val)})
	}
	return val, err
}

func (r *RecordingPrompter) AskText(question string, defaultVal string) (string, error) {
	val, err := r.Inner.AskText(question, defaultVal)
	if err == nil {
		r.Records = append(r.Records, QARecord{Question: question, Answer: val})
	}
	return val, err
}

func (r *RecordingPrompter) AskNumber(question string, defaultVal int) (int, error) {
	val, err := r.Inner.AskNumber(question, defaultVal)
	if err == nil {
		r.Records = append(r.Records, QARecord{Question: question, Answer: strconv.Itoa(val)})
	}
	return val, err
}

func (r *RecordingPrompter) ShowInferred(question string, value string) (bool, error) {
	accepted, err := r.Inner.ShowInferred(question, value)
	if err == nil {
		action := "changed"
		if accepted {
			action = "accepted"
		}
		r.Records = append(r.Records, QARecord{Question: question, Answer: fmt.Sprintf("[inferred: %s] %s", value, action)})
	}
	return accepted, err
}
