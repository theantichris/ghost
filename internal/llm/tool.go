package llm

// ToolProperty describes a single parameter that a tool accepts.
type ToolProperty struct {
	Type        string `json:"type"`        // Type of the property.
	Description string `json:"description"` // Description of the property.
}
