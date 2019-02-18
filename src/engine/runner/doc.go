// Package runner is responsible for looking at the state of the current node,
// and then ensuring that the information is replicated to match.
//
// The most common use for this is setting up node specific properties, and then
// having them replicated throughout the gossip protocol. These include things
// such as setting up swap, firewall, creating directories, files, and bricks.
package runner
