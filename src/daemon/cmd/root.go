package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"orbit.sh/engine"
)

var rootCmd = &cobra.Command{
	Use:   "orbitd",
	Short: "The primary runtime for the Orbit engine",
	Run: func(cmd *cobra.Command, args []string) {
		engine.Start()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
