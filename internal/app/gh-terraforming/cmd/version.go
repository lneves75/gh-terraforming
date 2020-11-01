package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionString = "0.0.1"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gh-terraforming",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gh-terraforming version", versionString)
	},
}
