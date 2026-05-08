package core

import (
	"encoding/json"
	"maps"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Session represents the active state of the User's current working context.
// serialized to ~/.local/state/rince/session.toml
type Session struct {
	// Unique ID representing the session
	ID string `toml:"id"`

	// Last Time Accessed which looks for stale Meta states and confirms with User
	LastAccessed time.Time `toml:"last_accessed"`

	// Subject is the topmost classification of a note. A note can be without a meta tag,
	// but cannot be without a Subject
	Subject string `toml:"subject"`

	// SubjectID helps with searching through Database.
	SubjectID string `toml:"subject_id, omitempty"`

	// Meta is the single source of true meta tags for the current context
	Meta map[string]string `toml:"meta"`

	// Tracks the current leaf node of the graph in the context.
	// e.g. If in Chapter, this holds Chapter's Node ID.
	CurrentParentID string `toml:"current_parent_id, omitempty"`

	// LastBatchID acts as a flag for the `undo` command.
	// If a log was created during the last state change, this holds its UUID
	// so the CLI knows it must execute a deep rollback (DB + Ledger + Session).
	LastBatchID string `toml:"last_log_id, omitempty"`
}

// SaveToTOML is also changed to Save() for the SessionManager

// SnapshotMeta method creates a deep copy of both the Subject and the Meta fields.
func (s *Session) SnapshotMeta() map[string]string {
	if s.Meta == nil {
		return make(map[string]string)
	}

	snapshot := make(map[string]string, len(s.Meta))
	maps.Copy(snapshot, s.Meta)
	return snapshot
}

// Backup and Undo commands are given to the Manager of the Session state.

// Called to show a marshalled representation in a JSON format
// for all fields in current session
func (s *Session) ToJSON() (string, error) {
	tomlBytes, err := toml.Marshal(s)
	if err != nil {
		return "", err
	}

	var intermediate map[string]any
	err = toml.Unmarshal(tomlBytes, &intermediate)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(intermediate, "", " ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Methods for Session Manipulation

// Set Subject for the Session
func (s *Session) SetSubject(subject string) {
	s.Subject = subject
}

// SetMeta allows setting or updating a meta tag in the Session's Meta map.
func (s *Session) SetMeta(key, value string) {
	if s.Meta == nil {
		s.Meta = make(map[string]string)
	}
	s.Meta[key] = value
}

// ResetMeta allows removing a meta tag from the Session's Meta map.
func (s *Session) ResetMeta(key string) {
	delete(s.Meta, key)
}

// Set Last Access time to now. To be called when creating a note.
func (s *Session) UpdateLastAccessed() {
	s.LastAccessed = time.Now()
}

// Non Methods and Loading Functions

// NewSession creates and initializes a new Session instance.
func NewSession(id, subject string) *Session {
	return &Session{
		ID:           id,
		LastAccessed: time.Now(),
		Subject:      subject,
		Meta:         make(map[string]string),
		LastBatchID:  "",
	}
}
