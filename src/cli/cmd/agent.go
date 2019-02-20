package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Socket is the UNIX socket to listen for agent requests on.
	Socket string
	// Port is the TCP port to listen for agent requests on.
	Port int
	// EnableRemoteAPI is whether or not to listen for TCP agent requests. This is
	// often used in conjunction with the Port variable to determine whether or
	// not remote API requests can be made.
	EnableRemoteAPI bool
	// RaftPort is the port used for Raft communication.
	RaftPort int
	// SerfPort is the port used for LAN serf communication.
	SerfPort int
	// WANSerfPort is the port used for WAN serf federation.
	WANSerfPort int
)

func init() {
	agentCmd.Flags().StringVarP(&Socket, "socket", "s", "/var/run/orbit.sock", "unix socket to listen to agent requests on")
	agentCmd.Flags().IntVarP(&Port, "port", "p", 6501, "port to listen to agent requests on")
	agentCmd.Flags().BoolVarP(&EnableRemoteAPI, "remote-api", "r", false, "enable agent requests on the port flag")
	agentCmd.Flags().IntVar(&RaftPort, "raft-port", 6502, "port to use for raft communication")
	agentCmd.Flags().IntVar(&SerfPort, "serf-port", 6503, "port to use for serf communication")
	agentCmd.Flags().IntVar(&WANSerfPort, "wan-serf-port", 6504, "port to use for multi-cluster WAN federation")

	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the primary long-running background process",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
