package llm

import "context"

// MockLLMClient mocks the LLMClient interface for testing.
type MockLLMClient struct {
	Error        error
	GenerateFunc func(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	VersionFunc  func(ctx context.Context) (string, error)
}

// Generate mocks the generate function.
func (llm MockLLMClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if llm.GenerateFunc != nil {
		return llm.GenerateFunc(ctx, systemPrompt, userPrompt)
	}

	if llm.Error != nil {
		return "", llm.Error
	}

	return "", nil
}

// Version gets the installed version of Ollama.
func (llm MockLLMClient) Version(ctx context.Context) (string, error) {
	if llm.VersionFunc != nil {
		return llm.VersionFunc(ctx)
	}

	if llm.Error != nil {
		return "", llm.Error
	}

	return "", nil
}
