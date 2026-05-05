package command

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rincewind_log/internal/core"
	"github.com/rincewind_log/internal/parser" // Your new Lexer package
	"github.com/spf13/cobra"
)

var writeNote = &cobra.Command{
	Use:   "note",
	Short: "Write notes in Text editor",
	Long:  "Opens Text Editor to write long form notes",
	Run:   runNote,
}

func init() {
	rootCmd.AddCommand(writeNote)
}

var Files core.FileStruct

func runNote(cmd *cobra.Command, args []string) {
	// 1. Setup temporary file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("rince-%d.txt", time.Now().Unix()))
	os.WriteFile(tmpFile, []byte(""), 0644)
	defer os.Remove(tmpFile) // Ensure cleanup happens automatically

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	// 2. Launch Editor
	execCmd := exec.Command(editor, tmpFile)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	err := execCmd.Run()
	if err != nil {
		fmt.Printf("Editor couldn't be opened: %v\n ", err)
		return
	}

	crudeContent, err := os.ReadFile(tmpFile)
	if err != nil || len(strings.TrimSpace(string(crudeContent))) == 0 {
		fmt.Printf("Log wasn't saved or was aborted: %v\n", err)
		return
	}

	// ==========================================
	// THE NEW LEXER PIPELINE
	// ==========================================

	// 3. Generate the Batch ID for this specific Neovim session
	batchID := uuid.New().String()

	// 4. Pull context from the active session
	// (Assuming SessionManager exposes Subject and Meta map)
	subject := ProgSession.ActiveState.Subject
	activeMeta := ProgSession.ActiveState.Meta

	// 5. Initialize the Lexer state machine
	expander := parser.NewExpander(subject, activeMeta, batchID)

	// 6. Parse the raw text into structured LogEntry DAOs
	entries := expander.Parse(string(crudeContent))

	if len(entries) == 0 {
		fmt.Println("No valid notes found to save. Did you forget your triggers (n, td, q)?")
		return
	}

	// 7. Pass the structured data to the database layer
	saveEntries(entries)
}

// saveEntries replaces the old flat-file writing logic.
// It receives perfectly structured, type-safe structs ready for SQLite.
func saveEntries(entries []parser.LogEntry) {
	// TODO: Handoff to SQLite Database package.
	// e.g., database.InsertNotes(entries)

	// For now, we will just simulate the database layer to prove the Lexer works:
	fmt.Printf("\n✓ Successfully parsed %d entries.\n", len(entries))
	fmt.Println(strings.Repeat("-", 40))

	for _, entry := range entries {
		fmt.Printf("ID:       %s\n", entry.ID)
		fmt.Printf("Batch:    %s (Seq: %d)\n", entry.BatchID, entry.Sequence)
		fmt.Printf("Type:     %s\n", entry.Type)
		fmt.Printf("Subject:  %s\n", entry.Subject)
		fmt.Printf("Book:     %s\n", entry.Meta["book"])
		fmt.Printf("Chapter:  %s\n", entry.Meta["chapter"])
		fmt.Printf("Content:  %s...\n", strings.Split(entry.Content, "\n")[0]) // Print just the first line
		fmt.Println(strings.Repeat("-", 40))
	}

	log.Printf("Simulated save for %d entries.\n", len(entries))
}
