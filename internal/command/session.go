package command

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Prints the current session state",
	Long:  "By default, shows current session state. Other subcommands follow, such as show and set",
	Run:   runSessionView,
}

// Add subcommand to sessionCmd to show session details in JSON - rince session show
var sessionShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current session state in JSON format",
	Long:  "Loads the active session from the TOML file and prints its contents as a JSON object.",
	Run:   runSessionView,
}

var subjectCmd = &cobra.Command{
	Use:   "subject",
	Short: "Show the current session subject, by default.",
	Long:  "By default, show the current session subject.\n rince <object> <action> \nAfter this, follow with <action> for subjects",
	Run:   runSubjectView,
}

var subjectShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current session subject",
	Long:  "Loads the active session from the TOML file and prints its subject.",
	Run:   runSubjectView,
}

var subjectSetCmd = &cobra.Command{
	Use:   "set [subject]",
	Short: "Set the current session subject",
	Long:  "Sets the current session subject to the provided value and saves the session state.",
	Args:  cobra.MinimumNArgs(1),
	Run:   runSessionSetSubject,
}

func runSessionView(cmd *cobra.Command, args []string) {
	sessionJSON, err := ProgSession.ActiveState.ToJSON()
	if err != nil {
		slog.Error("Converting session to JSON", "error", err.Error())
		return
	}
	fmt.Println(sessionJSON)
}

func runSubjectView(cmd *cobra.Command, args []string) {
	// Since subject always exists from PersistentPreRun,
	fmt.Printf("Current subject: %s\n", ProgSession.ActiveState.Subject)
}

func runSessionSetSubject(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Please provide a subject to set.")
		return
	}
	subjectName := strings.Join(args, " ")

	// Before setting subject, backup current state.
	ProgSession.BackupState() // No log ID since this is just a state change

	ProgSession.ActiveState.SetSubject(subjectName)
	err := ProgSession.Save()
	if err != nil {
		slog.Error("Saving session after setting subject", "error", err.Error())
		return
	}
	slog.Info("Subject updated", "new_subject", subjectName)
	fmt.Printf("Subject set to: %s\n", subjectName)
}

func init() {
	sessionCmd.AddCommand(sessionShowCmd)

	subjectCmd.AddCommand(subjectShowCmd)
	subjectCmd.AddCommand(subjectSetCmd)

	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(subjectCmd)
}
