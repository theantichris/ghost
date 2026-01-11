package llm

// ToolParameters defines all the parameters a tool accepts. It wraps multiple ToolProperty values and specifies which ones are required.
type ToolParameters struct {
	Type       string                  `json:"type"`
	Required   []string                `json:"required"`
	Properties map[string]ToolProperty `json:"properties"`
}

// ToolProperty describes a single parameter that a tool accepts.
type ToolProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
