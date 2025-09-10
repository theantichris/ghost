package llm

// Define ollama client struct
type OllamaClient struct {
	baseURL      string // Base Ollama server URL
	defaultModel string // Default model to use
}

// Instantiate ollama llm client
func NewOllamaClient(baseURL, defaultModel string) *OllamaClient {
	return &OllamaClient{
		baseURL:      baseURL,
		defaultModel: defaultModel,
	}
}
