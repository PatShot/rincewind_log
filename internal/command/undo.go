package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last session state change",
	Long:  "Reverts the current session's meta tags to their previous state.",
	Run: func(cmd *cobra.Command, args []string) {

		batchID, err := ProgSession.Undo()
		if err != nil {
			fmt.Println(err)
			return
		}

		// err = database.DeleteBatch(batchID)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		fmt.Printf("Undone Batch ID :: %s", batchID)

		err = ProgSession.Save()
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
