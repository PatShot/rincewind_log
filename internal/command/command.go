// Package command is the organization module for the main CLI commands
package command

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rincewind_log/internal/core"
	"github.com/rincewind_log/internal/database"
	"github.com/spf13/cobra"
)

var ProgSession *core.SessionManager

// quick list of commands that require db access.
var requireDbCmds = map[string]bool{
	"log":  true,
	"note": true,
}

var rootCmd = &cobra.Command{
	Use:   "rince",
	Short: "A CLI log to dump typed notes",
	Long:  "Rincewind logs notes to a Database and a Log File, which can then be synced as required and formatted as required.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmdName := cmd.Name()

		ifExempt := cmdName == "config" || cmdName == "help" || cmdName == "version" || cmdName == "subject" || cmdName == "session"

		if ProgSession.ActiveState.Subject == "" && !ifExempt {
			slog.Error("Active Subject is an empty string!", "error", "[BLANK SUBJECT]")
			fmt.Println("No active subject. Please run `rince subject set [subject]` to set an active subject.")
			os.Exit(1)
		}

		ProgSession.ActiveState.UpdateLastAccessed()

		logfile, err := resolveLogFilePath()
		if err != nil {
			slog.Error("Failed to resolve log file path", "error", err.Error())
		}
		Files.SetLogFileName(logfile)

		mainfile, err := resolveMainFilePath()
		if err != nil {
			slog.Error("Failed to resolve main file path", "error", err.Error())
		}
		Files.SetMainFileName(mainfile)

		dbfile, err := resolveDBPath()
		if err != nil {
			slog.Error("Failed to resolve database file path", "error", err.Error())
		}
		Files.SetDBFileName(dbfile)

		if requireDbCmds[cmd.Name()] {
			err = database.InitDB(Files.DBFile)
		}

		return err
	},
}

// func init() {
// 	// Quick Dirty Config with Tests
// 	// FUTUREWORK: Implement a proper config system, and move this to a more appropriate place.
// 	var configLogFile = "/home/patg/Projects/study-log/rince_test.log"
// 	var configMainFile = "/home/patg/Projects/study-log/rince_main.md"
// 	var dbFile = "/home/patg/Projects/study-log/rincewind.db"
// 	Files.SetLogFileName(configLogFile)
// 	Files.SetMainFileName(configMainFile)
// 	Files.SetDBFileName(dbFile)
// }

func resolveDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(home, ".rincewind", "data.db"), nil
	case "darwin":
		return filepath.Join(home, ".rincewind", "data.db"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func resolveMainFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(home, ".rincewind", "main.md"), nil
	case "darwin":
		return filepath.Join(home, ".rincewind", "main.md"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func resolveLogFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(home, ".rincewind", "rincelog.log"), nil
	case "darwin":
		return filepath.Join(home, ".rincewind", "rincelog.log"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func Execute() {
	// Execute is entrypoint for Cobra.
	// Any config adjustments should be done here.
	// It is definitely more idiomatic to save in .local and not in .rincewind.
	// I guess I'll have this for now.
	// FUTUREWORK: Implement a proper config system, and move this to a more appropriate place.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}

	// This should come from the Config File?
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
