package llm

import "context"

// MockLLMClient mocks the LLMClient interface for testing purposes.
// Set GenerateFunc, VersionFunc, or ShowFunc to provide custom implementations.
// Set Error to have all methods return that error by default.
type MockLLMClient struct {
	Error        error
	GenerateFunc func(ctx context.Context, systemPrompt, userPrompt string, images []string) (string, error)
	VersionFunc  func(ctx context.Context) (string, error)
	ShowFunc     func(ctx context.Context, model string) error
}

// Generate mocks the Generate method by calling GenerateFunc if set, returning Error if set, or returning an empty string.
func (llm MockLLMClient) Generate(ctx context.Context, systemPrompt, userPrompt string, images []string) (string, error) {
	if llm.GenerateFunc != nil {
		return llm.GenerateFunc(ctx, systemPrompt, userPrompt, images)
	}

	if llm.Error != nil {
		return "", llm.Error
	}

	return "", nil
}

// Version mocks the Version method by calling VersionFunc if set, returning Error if set, or returning an empty string.
func (llm MockLLMClient) Version(ctx context.Context) (string, error) {
	if llm.VersionFunc != nil {
		return llm.VersionFunc(ctx)
	}

	if llm.Error != nil {
		return "", llm.Error
	}

	return "", nil
}

// Show mocks the Show method by calling ShowFunc if set, returning Error if set, or returning nil.
func (llm MockLLMClient) Show(ctx context.Context, model string) error {
	if llm.ShowFunc != nil {
		return llm.ShowFunc(ctx, model)
	}

	if llm.Error != nil {
		return llm.Error
	}

	return nil
}
