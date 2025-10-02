package flags

// FlagSetRequest toggles a single feature flag.
type FlagSetRequest struct {
	Flag  string `json:"flag"`
	Value bool   `json:"value"`
}

// FlagAssignments represents a collection of flag states keyed by flag name.
type FlagAssignments map[string]bool

// FlagUnsetRequest identifies the flag to remove.
type FlagUnsetRequest struct {
	Flag string `json:"flag"`
}
