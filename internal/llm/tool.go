package llm

// Tool defines a tool available to the LLM.
type Function struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a callable function for the LLM.
type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

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

// ToolCall represents the LLM's request to invoke a tool.
type ToolCall struct {
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}
