package command

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/rincewind_log/internal/database"
	"github.com/rincewind_log/internal/parser"
	"github.com/spf13/cobra"
)

var logNoteCmd = &cobra.Command{
	Use:   "log [string]",
	Short: "Logs a string to the log file and database",
	Long:  "Logs a string provided to the database and log file, along with various meta fields, stored under a Subject. By default, the subject is INIT.",
	Args:  cobra.ExactArgs(1),
	Run:   runLogNote,
}

// runLogNote takes argument, provides to parser, then writes to file and database.
// In general, it follows how note command proceeds, without opening the editor.
// It must also update the session's PreviousState and LastLogID for undo functionality, and then save the session.
func runLogNote(cmd *cobra.Command, args []string) {
	noteContent := args[0]

	// 1. Generate the Batch ID for this specific transaction
	batchID := uuid.New().String()

	// 2. Safely pull context from the active session
	subject := ProgSession.ActiveState.Subject
	if subject == "" {
		subject = "INIT"
	}

	activeMeta := ProgSession.ActiveState.Meta
	if activeMeta == nil {
		activeMeta = make(map[string]string)
	}

	// 3. Initialize the Lexer
	expander := parser.NewExpander(subject, activeMeta, batchID)

	// 4. Parse the raw string into structured LogEntry DAOs
	entries := expander.Parse(noteContent)

	if len(entries) == 0 {
		fmt.Println("No valid notes found. Did you forget a trigger (like 'n ', 'td ', or 'q ')?")
		return
	}

	// 5. Save to the database
	// Just Debug
	saveEntries(entries)
	newSubjectID, newParentId, err := database.InsertNode(
		entries,
		ProgSession.ActiveState.SubjectID,
		ProgSession.ActiveState.CurrentParentID,
	)
	if err != nil {
		slog.Error("Failed to save to Database", "error", err)
		return
	}

	// 6. Session Management & Undo State
	// Back up the current state before making changes (assuming a simple struct copy works here)
	ProgSession.BackupState()

	// Track the BatchID so `rince undo` can delete all rows associated with this command execution
	ProgSession.ActiveState.LastBatchID = batchID

	ProgSession.ActiveState.SubjectID = newSubjectID
	ProgSession.ActiveState.CurrentParentID = newParentId

	// Update the active session's Meta so subsequent commands remember where we left off.
	// If the user typed "b NewBook n Note", we want "NewBook" to persist.
	lastEntry := entries[len(entries)-1]
	ProgSession.ActiveState.Meta["book"] = lastEntry.Meta["book"]
	ProgSession.ActiveState.Meta["chapter"] = lastEntry.Meta["chapter"]

	// 7. Save the updated Session back to disk/db
	err = ProgSession.Save()
	if err != nil {
		slog.Error("Failed to save session state", "error", err)
		fmt.Println("Note saved, but session state could not be updated.")
		return
	}

	fmt.Printf("Successfully logged %d entry(s) under batch %s\n", len(entries), batchID)
}

func init() {
	rootCmd.AddCommand(logNoteCmd)
}
