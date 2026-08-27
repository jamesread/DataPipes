package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jamesread/data-cleaner/internal/config"
)

const maxExecutionsPerJob = 200

// Entry is a persisted pipeline execution record.
type Entry struct {
	ID          string `json:"id"`
	JobID       string `json:"job_id"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
	RunKind     string `json:"run_kind"`
	StepDetail  string `json:"step_detail"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	TriggeredBy string `json:"triggered_by"`
}

type fileData struct {
	Executions []Entry `json:"executions"`
}

// Store persists execution history to a JSON file.
type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore() *Store {
	return &Store{path: historyFilePath()}
}

func historyFilePath() string {
	configPath, err := config.LoadStatus()
	if err == nil && configPath != "" && configPath != "sample-config.yaml (built-in)" {
		return filepath.Join(filepath.Dir(configPath), "executions.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "executions.json"
	}
	return filepath.Join(home, ".datapipes", "executions.json")
}

func (s *Store) Append(entry Entry) error {
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	if entry.TriggeredBy == "" {
		entry.TriggeredBy = "manual"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readLocked()
	if err != nil {
		return err
	}

	data.Executions = append(data.Executions, entry)
	data.Executions = trimForJob(data.Executions, entry.JobID, maxExecutionsPerJob)

	return s.writeLocked(data)
}

func trimForJob(entries []Entry, jobID string, limit int) []Entry {
	if limit <= 0 {
		return entries
	}
	var jobEntries []Entry
	var other []Entry
	for _, e := range entries {
		if e.JobID == jobID {
			jobEntries = append(jobEntries, e)
		} else {
			other = append(other, e)
		}
	}
	if len(jobEntries) > limit {
		jobEntries = jobEntries[len(jobEntries)-limit:]
	}
	return append(other, jobEntries...)
}

func (s *Store) ListJob(jobID string) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readLocked()
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0)
	for _, e := range data.Executions {
		if e.JobID == jobID {
			out = append(out, e)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt > out[j].StartedAt
	})

	return out, nil
}

func (s *Store) readLocked() (fileData, error) {
	if s.path == "" {
		return fileData{}, fmt.Errorf("history path is empty")
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileData{Executions: []Entry{}}, nil
		}
		return fileData{}, err
	}
	var data fileData
	if err := json.Unmarshal(raw, &data); err != nil {
		return fileData{}, err
	}
	if data.Executions == nil {
		data.Executions = []Entry{}
	}
	return data, nil
}

func (s *Store) writeLocked(data fileData) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(s.path, raw, 0o644)
}

// Record builds and stores an execution entry.
func (s *Store) Record(jobID, runKind, stepDetail string, started time.Time, success bool, message string) error {
	if jobID == "" {
		jobID = config.DefaultJobID
	}
	completed := time.Now().UTC()
	return s.Append(Entry{
		JobID:       jobID,
		StartedAt:   started.UTC().Format(time.RFC3339),
		CompletedAt: completed.Format(time.RFC3339),
		RunKind:     runKind,
		StepDetail:  stepDetail,
		Success:     success,
		Message:     message,
		TriggeredBy: "manual",
	})
}
