// Package database provides functions for interacting with the database.
package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rincewind_log/internal/core"
	"github.com/rincewind_log/internal/parser"
)

// Global connetion pool
var DB *sql.DB

// InitDB to initialize
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if it doesn't exist
	_, err = DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes
	_, err = DB.Exec(index_creation)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	slog.Debug("Database initialized successfully", "path", dbPath)
	return nil
}

// ensureNode inserts a node under a parent if the node doesn't exist, and if already exists, then does nothing.
// returns the node ID of the existing or newly created node.
func ensureNode(tx *sql.Tx, parentID *string, nodeType, label, batchID string) (string, error) {
	var id string

	query := `SELECT id FROM nodes WHERE node_type = ? and label = ? and parent_id IS ?`
	err := tx.QueryRow(query, nodeType, label, parentID).Scan(&id)

	if err == sql.ErrNoRows {
		id = uuid.New().String()
		insertQuery := `INSERT INTO nodes (id, parent_id, batch_id, sequence, node_type, label) VALUES (?, ?, ?, 0, ?, ?)`
		_, err = tx.Exec(insertQuery, id, parentID, batchID, nodeType, label)
		if err != nil {
			return "", fmt.Errorf("failed to create %snode: %s: %w", nodeType, label, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to query for %snode: %s: %w", nodeType, label, err)
	}
	return id, nil
}

// InsertNode inserts a new node into the database and returns its ID.
func InsertNode(entries []parser.LogEntry, activeSubjectID string, activeParentID string) (string, string, error) {
	if len(entries) == 0 {
		return activeSubjectID, activeParentID, nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return "", "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// FUTUREWORK: Adapt GetHierarchy to change with GetHierarchy(name string)
	// Multiple hierarchies should be supported, if in case that has been set by the user.
	// For now, we will just use the default hierarchy.
	hierarchy := core.GetHierarchy()

	var finalSubjectID string
	var finalParentID string

	for _, entry := range entries {
		var currentParentID *string = nil
		var nodeID string

		// 1. Ensure Subject Node
		// Ensure that the node exists under the "subject", which is the root node.
		// That makes the nodeID gotten the current ParentID in the hierarchy.
		nodeID, err = ensureNode(tx, currentParentID, "subject", entry.Subject, entry.BatchID)
		if err != nil {
			return "", "", fmt.Errorf("failed to ensure subject node: %w", err)
		}
		currentParentID = &nodeID
		finalSubjectID = nodeID

		// 2. Traverse / Create the Meta nodes (Level 2+)
		for _, metaType := range hierarchy {
			if label, exists := entry.Meta[metaType]; exists {
				nodeID, err = ensureNode(tx, currentParentID, metaType, label, entry.BatchID)
				if err != nil {
					return "", "", fmt.Errorf("failed to ensure %s node: %w", metaType, err)
				}
				currentParentID = &nodeID
			}
		}

		// Final ParentID will be the last node found in the Meta that corresponds to the hierarchy.
		// It may very well be that the a node in the hierarchy is not found. In that case, the hierarchy gets skipped.
		// As an example, consider the Hierarchy: topic -> subtopic -> category -> article.
		// If the topic doens't have a subtopic, but has a category and article associated with it,
		// then the category and article will be created under the topic, and the subtopic level will be skipped.
		finalParentID = *currentParentID

		// 3. Add the Leaf Node
		noteID := uuid.New().String()
		insertQuery := `INSERT INTO nodes (id, parent_id, batch_id, sequence, node_type, content) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.Exec(insertQuery, noteID, currentParentID, entry.BatchID, entry.Sequence, entry.Type, entry.Content)
		if err != nil {
			return "", "", fmt.Errorf("failed to insert note node: %w", err)
		}

	}
	return finalSubjectID, finalParentID, nil
}
