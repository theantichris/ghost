package llm

import "context"

// MockLLMClient mocks the LLMClient interface for testing.
type MockLLMClient struct {
	Error        error
	GenerateFunc func(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	VersionFunc  func(ctx context.Context) (string, error)
	ShowFunc     func(ctx context.Context) error
}

// Generate mocks the Generate function.
func (llm MockLLMClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if llm.GenerateFunc != nil {
		return llm.GenerateFunc(ctx, systemPrompt, userPrompt)
	}

	if llm.Error != nil {
		return "", llm.Error
	}

	return "", nil
}

// Version mocks the Version function.
func (llm MockLLMClient) Version(ctx context.Context) (string, error) {
	if llm.VersionFunc != nil {
		return llm.VersionFunc(ctx)
	}

	if llm.Error != nil {
		return "", llm.Error
	}

	return "", nil
}

// Show mocks the Show function.
func (llm MockLLMClient) Show(ctx context.Context) error {
	if llm.ShowFunc != nil {
		return llm.ShowFunc(ctx)
	}

	if llm.Error != nil {
		return llm.Error
	}

	return nil
}
