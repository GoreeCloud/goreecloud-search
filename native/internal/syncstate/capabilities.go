package syncstate

// Capability is the stable GoreeCloud Sync-facing description of a Search
// dataset. Storage details stay private to Search; only protocol capabilities
// cross the application boundary.
type Capability struct {
	Dataset       string `json:"dataset"`
	Application   string `json:"application"`
	SchemaVersion int    `json:"schemaVersion"`
	Read          bool   `json:"read"`
	Write         bool   `json:"write"`
	Delete        bool   `json:"delete"`
}

var capabilities = []Capability{
	{Dataset: "search.preferences", Application: "search", SchemaVersion: 1, Read: true, Write: true},
	{Dataset: "search.history", Application: "search", SchemaVersion: 1, Read: true, Write: true, Delete: true},
	{Dataset: "search.sources", Application: "search", SchemaVersion: 1, Read: true, Write: true},
}

// Capabilities returns an isolated copy safe for API serialization.
func Capabilities() []Capability {
	return append([]Capability(nil), capabilities...)
}
