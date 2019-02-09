package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "orbitd",
	Short: "The primary runtime for the Orbit engine",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Now starting the orbit engine...")
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
