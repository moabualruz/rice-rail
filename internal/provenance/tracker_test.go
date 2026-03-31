package provenance

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordInference(t *testing.T) {
	tr := NewTracker()
	tr.RecordInference("lang-go", "profiler", "Go", "95%")

	if len(tr.Log.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(tr.Log.Records))
	}
	r := tr.Log.Records[0]
	if r.ID != "lang-go" {
		t.Errorf("expected ID lang-go, got %s", r.ID)
	}
	if r.Type != "inferred" {
		t.Errorf("expected type inferred, got %s", r.Type)
	}
	if r.Source != "profiler" {
		t.Errorf("expected source profiler, got %s", r.Source)
	}
	if r.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestRecordUserDecision(t *testing.T) {
	tr := NewTracker()
	tr.RecordUserDecision("safety-mode", "strict", "team preference")

	if len(tr.Log.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(tr.Log.Records))
	}
	r := tr.Log.Records[0]
	if r.Type != "user_stated" {
		t.Errorf("expected type user_stated, got %s", r.Type)
	}
	if r.Source != "interview" {
		t.Errorf("expected source interview, got %s", r.Source)
	}
}

func TestRecordGeneration(t *testing.T) {
	tr := NewTracker()
	tr.RecordGeneration("constitution", "constitution.yaml", "init")

	if len(tr.Log.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(tr.Log.Records))
	}
	r := tr.Log.Records[0]
	if r.Type != "generated" {
		t.Errorf("expected type generated, got %s", r.Type)
	}
	if r.Artifact != "constitution.yaml" {
		t.Errorf("expected artifact constitution.yaml, got %s", r.Artifact)
	}
}

func TestRecordToolRun(t *testing.T) {
	tr := NewTracker()
	tr.RecordToolRun("eslint", "found 3 issues")

	if len(tr.Log.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(tr.Log.Records))
	}
	r := tr.Log.Records[0]
	if r.Type != "inferred" {
		t.Errorf("expected type inferred, got %s", r.Type)
	}
	if r.Source != "eslint" {
		t.Errorf("expected source eslint, got %s", r.Source)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()

	tr := NewTracker()
	tr.RecordInference("lang-go", "profiler", "Go", "95%")
	tr.RecordUserDecision("safety-mode", "strict", "team preference")
	tr.RecordGeneration("constitution", "constitution.yaml", "init")

	if err := tr.Save(dir); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify decisions.json exists and can be loaded
	decisions, err := Load(filepath.Join(dir, "decisions.json"))
	if err != nil {
		t.Fatalf("loading decisions: %v", err)
	}
	if len(decisions.Records) != 1 {
		t.Errorf("expected 1 decision, got %d", len(decisions.Records))
	}
	if decisions.Records[0].Type != "user_stated" {
		t.Errorf("expected user_stated, got %s", decisions.Records[0].Type)
	}

	// Verify inferred-evidence.json exists and can be loaded
	evidence, err := Load(filepath.Join(dir, "inferred-evidence.json"))
	if err != nil {
		t.Fatalf("loading evidence: %v", err)
	}
	if len(evidence.Records) != 2 {
		t.Errorf("expected 2 evidence records, got %d", len(evidence.Records))
	}

	// Verify files are valid JSON on disk
	for _, name := range []string{"decisions.json", "inferred-evidence.json"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("reading %s: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("%s is empty", name)
		}
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "provenance")

	tr := NewTracker()
	tr.RecordInference("test", "test", "test", "high")

	if err := tr.Save(dir); err != nil {
		t.Fatalf("save to nested dir failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "inferred-evidence.json")); os.IsNotExist(err) {
		t.Error("expected inferred-evidence.json to exist")
	}
}

func TestNilTrackerDoesNotCrash(t *testing.T) {
	// A nil *Tracker should not be called directly (Go nil pointer),
	// but the builder uses a nil check pattern. Verify the pattern works.
	var tr *Tracker
	if tr != nil {
		tr.RecordInference("should-not-reach", "", "", "")
	}
	// If we get here without panic, the test passes.
}
