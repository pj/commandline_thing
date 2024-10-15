package pkg

// These three types should uniquely identify an operation instance, the kind of location config we are generating,
// and the specific instance of that location and finally the operation within that location.

// LocationKey is the type of thing we are generating e.g. a pane status, prompt, etc
type LocationKey string

// The specific instance of a location we are generating/updating/storing for, usually a specific tmux window or pane.
type InstanceKey string

// Name of the operation e.g. "branch", "git", etc
type OperationName string
