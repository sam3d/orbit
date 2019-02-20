package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"orbit.sh/engine"
)

var rootCmd = &cobra.Command{
	Use:   "orbitd",
	Short: "The primary runtime for the Orbit engine",
	Run: func(cmd *cobra.Command, args []string) {
		engine.Start()

		// Allow graceful exit
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGINT)
		<-exit
		fmt.Println("\nExiting...")
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
