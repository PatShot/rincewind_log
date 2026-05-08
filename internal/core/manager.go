package core

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
)

// SessionState represents the backup for Subject and Meta before state changes
type StateSnapshot struct {
	Subject         string            `toml:"subject"`
	SubjectID       string            `toml:"subject_id"`
	Meta            map[string]string `toml:"meta"`
	CurrentParentID string            `toml:"current_parent_id"`
}

// SessionManager is the orchestrator, which alongside the PersistentPreRun in the root,
// manages Session for the duration of the program run.

type SessionManager struct {
	ActiveState   *Session       `toml:"active_state"`
	PreviousState *StateSnapshot `toml:"previous_state, omitempty"`
	// Internal, therefore excluded from toml save
	sessionPath string `toml:"-"`
}

// Save marshals the SessionStates to a TOML file.
func (m *SessionManager) Save() error {
	data, err := toml.Marshal(m)
	if err != nil {
		return err
	}

	err = os.WriteFile(m.sessionPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Backup the State. To be used before modification of any kind.
func (m *SessionManager) BackupState() {
	if m.ActiveState == nil {
		return
	}

	m.PreviousState = &StateSnapshot{
		Subject:         m.ActiveState.Subject,
		Meta:            maps.Clone(m.ActiveState.Meta),
		SubjectID:       m.ActiveState.SubjectID,
		CurrentParentID: m.ActiveState.CurrentParentID,
	}
}

// Undo the ActiveState and load from the saved PreviousState
// Returns a batchID that is the string representing a batch of operations
// to be undone in the SQL database.
func (m *SessionManager) Undo() (string, error) {
	if m.PreviousState == nil {
		return "", fmt.Errorf("There is no previous state")
	}

	// Must capture the Batch ID now
	batchIDtoRollback := m.ActiveState.LastBatchID
	if batchIDtoRollback == "" {
		return "", fmt.Errorf("No batchID found to rollback")
	}

	m.ActiveState.Subject = m.PreviousState.Subject
	m.ActiveState.Meta = maps.Clone(m.PreviousState.Meta)
	m.ActiveState.SubjectID = m.PreviousState.SubjectID
	m.ActiveState.CurrentParentID = m.PreviousState.CurrentParentID

	// Wipe LastBatchID. Do nothing, as LastBatchID provided elsewhere before action.
	m.ActiveState.LastBatchID = ""

	// Wipe PreviousState to not do double undo.
	m.PreviousState = nil

	// Save restored state to TOML

	return batchIDtoRollback, nil
}

func NewSessionManager(path string) (*SessionManager, error) {
	manager := &SessionManager{
		ActiveState:   nil,
		PreviousState: nil,
		sessionPath:   path,
	}

	// Ensure MkdirAll is called
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// If session file not present then create new session file.
	// This will only run once or twice, when user first starts or deletes the session file.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		manager.ActiveState = NewSession(uuid.New().String(), "INIT")
		if err := manager.Save(); err != nil {
			return nil, fmt.Errorf("failed to save new session: %w", err)
		}
	}

	// Load the session from file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	err = toml.Unmarshal(data, manager)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// If Meta wasn't read, Go will leave Meta as nil. This will lead to Panics
	// Initialize into a blank map.
	if manager.ActiveState != nil && manager.ActiveState.Meta == nil {
		manager.ActiveState.Meta = make(map[string]string)
	}

	if manager.ActiveState == nil {
		panic("ActiveState is nil")
	}

	return manager, nil
}
