// Package store provides an interface to the entire state of the distributed
// consensus algorithm. It is responsible for replicating changes and managing
// the underlying state of the orchestration. Reacting to those state changes is
// not the responsibility of the store, and takes place in the "runner" package.
package store
