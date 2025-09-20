package llm

// Role defines the role of a message in the chat.
type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
	Tool      Role = "tool"
)

// ChatRequest represents a request to the Ollama chat API.
type ChatRequest struct {
	Model    string        `json:"model"`    // Required. The model name.
	Messages []ChatMessage `json:"messages"` // The messages of the chat, this can be used to keep a chat memory
	Think    bool          `json:"think"`    // Whether to think step by step
	Stream   bool          `json:"stream"`   // Whether to stream the response
}

// ChatMessage represents a single message in the chat.
type ChatMessage struct {
	Role    Role   `json:"role"`    // The role of the message, either system, user, assistant, or tool
	Content string `json:"content"` // The content of the message
}

// ChatResponse represents a response from the Ollama chat API.
type ChatResponse struct {
	Message ChatMessage `json:"message"` // The response message from the assistant
}

// apiError represents an error response from the Ollama API.
type apiError struct {
	Error string `json:"error"`
}
