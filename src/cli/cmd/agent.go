package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"orbit.sh/engine"
)

var (
	// Socket is the UNIX socket to listen for agent requests on.
	Socket string
	// Port is the TCP port to listen for agent requests on.
	Port int
)

func init() {
	agentCmd.Flags().StringVarP(&Socket, "socket", "s", "/var/run/orbit.sock", "unix socket to listen to agent requests on ('' to disable)")
	agentCmd.Flags().IntVarP(&Port, "port", "p", 6505, "port to listen to agent requests on (-1 to disable)")

	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "Start the primary long-running background process",
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		// Configure the logger.
		log.SetFlags(log.LstdFlags)

		// Create the engine.
		e := engine.New()

		// Configure the API.
		e.APIServer.Port = Port
		e.APIServer.Socket = Socket

		// Start the engine.
		go func() {
			err := e.Start()
			log.Fatalf("fatal: %s\n", err)
		}()

		// Gracefully exit.
		exit := make(chan os.Signal)
		signal.Notify(exit, os.Interrupt)
		<-exit
		fmt.Println("\nReceived interrupt...")

		err := e.Stop()
		if err != nil {
			panic(err)
		}
	},
}
