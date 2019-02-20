package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "orbit",
	Short: "A simple and scalable self-hosted Platform as a Service",
}

// Execute will run the root cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
