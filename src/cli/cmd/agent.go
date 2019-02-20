package cmd

import (
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
	// RaftPort is the port used for Raft communication.
	RaftPort int
	// SerfPort is the port used for LAN serf communication.
	SerfPort int
	// WANSerfPort is the port used for WAN serf federation.
	WANSerfPort int
)

func init() {
	agentCmd.Flags().StringVarP(&Socket, "socket", "s", "/var/run/orbit.sock", "unix socket to listen to agent requests on ('' to disable)")
	agentCmd.Flags().IntVarP(&Port, "port", "p", 6501, "port to listen to agent requests on (-1 to disable)")
	agentCmd.Flags().IntVar(&RaftPort, "raft-port", 6502, "port to use for raft communication")
	agentCmd.Flags().IntVar(&SerfPort, "serf-port", 6503, "port to use for serf communication")
	agentCmd.Flags().IntVar(&WANSerfPort, "wan-serf-port", 6504, "port to use for multi-cluster WAN federation")

	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the primary long-running background process",
	Run: func(cmd *cobra.Command, args []string) {
		// Configure the logger.
		log.SetFlags(log.LstdFlags)

		// Create the engine.
		e := engine.New()

		// Configure the API.
		e.API.Port = Port
		e.API.Socket = Socket

		// Configure the store.
		e.Store.RaftPort = RaftPort
		e.Store.SerfPort = SerfPort
		e.Store.WANSerfPort = WANSerfPort

		// Start the engine.
		go func() {
			err := e.Start()
			if err != nil {
				log.Fatal(err)
			}
		}()

		// Gracefully exit.
		exit := make(chan os.Signal)
		signal.Notify(exit, os.Interrupt)
		<-exit
	},
}
