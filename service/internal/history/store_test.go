package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreAppendAndListJob(t *testing.T) {
	dir := t.TempDir()
	s := &Store{path: filepath.Join(dir, "executions.json")}

	started := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	if err := s.Record("job-a", "extract", "Extract only", started, true, ""); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := s.Record("job-b", "pipeline", "Pipeline (extract, transform, load)", started, false, "load failed"); err != nil {
		t.Fatalf("record: %v", err)
	}

	rows, err := s.ListJob("job-a")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].RunKind != "extract" || rows[0].TriggeredBy != "manual" {
		t.Fatalf("unexpected row: %+v", rows[0])
	}

	if _, err := os.Stat(s.path); err != nil {
		t.Fatalf("history file not written: %v", err)
	}
}
