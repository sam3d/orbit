package engine

// StoreState is the all-encompassing state of the cluster. The operations are
// performed to this after being cast to a finite state machine, and otherwise
// won't be able to make any changes.
//
// Important to note is that the state is not aware of its distributed nature,
// and is simply for keeping track of the current data.
type StoreState struct {
	Users Users `json:"users"`
}
