// Package command is the organization module for the main CLI commands
package command

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rincewind_log/internal/core"
	"github.com/spf13/cobra"
)

var ProgSession *core.SessionManager

var rootCmd = &cobra.Command{
	Use:   "rince",
	Short: "A CLI log to dump typed notes",
	Long:  "Rincewind logs notes to a Database and a Log File, which can then be synced as required and formatted as required.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmdName := cmd.Name()

		ifExempt := cmdName == "config" || cmdName == "help" || cmdName == "version" || cmdName == "subject" || cmdName == "session"

		if ProgSession.ActiveState.Subject == "" && !ifExempt {
			slog.Error("Active Subject is an empty string!", "error", "[BLANK SUBJECT]")
			fmt.Println("No active subject. Please run `rince subject set [subject]` to set an active subject.")
			os.Exit(1)
		}

		ProgSession.ActiveState.UpdateLastAccessed()
	},
}

var configLogFile = "/home/patg/Projects/study-log/rince_test.log"
var configMainFile = "/home/patg/Projects/study-log/rince_main.md"

func init() {
	Files.SetLogFileName(configLogFile)
	Files.SetMainFileName(configMainFile)
}

func Execute() {
	// Execute is entrypoint for Cobra.
	// Any config adjustments should be done here.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}

	sessionPath := filepath.Join(homeDir, ".rincewind", "session.toml")
	ProgSession, err = core.NewSessionManager(sessionPath)
	if err != nil {
		slog.Error("Failed to initialize session manager", "error", err.Error())
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
