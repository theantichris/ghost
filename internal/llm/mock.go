package llm

import "context"

// MockClient mocks the Client interface for testing purposes.
// Set GenerateFunc, VersionFunc, or ShowFunc to provide custom implementations.
// Set Error to have all methods return that error by default.
type MockClient struct {
	Error        error
	GenerateFunc func(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error
	VersionFunc  func(ctx context.Context) (string, error)
	ShowFunc     func(ctx context.Context, model string) error
}

// Generate mocks the client Generate method by calling GenerateFunc if set, returning Error if set, or calling callback with empty string.
func (llm MockClient) Generate(ctx context.Context, model, systemPrompt, userPrompt string, images []string, callback func(string) error) error {
	if llm.GenerateFunc != nil {
		return llm.GenerateFunc(ctx, model, systemPrompt, userPrompt, images, callback)
	}

	if llm.Error != nil {
		return llm.Error
	}

	return callback("")
}

// Version mocks the Version method by calling VersionFunc if set, returning Error if set, or returning an empty string.
func (llm MockClient) Version(ctx context.Context) (string, error) {
	if llm.VersionFunc != nil {
		return llm.VersionFunc(ctx)
	}

	if llm.Error != nil {
		return "", llm.Error
	}

	return "", nil
}

// Show mocks the Show method by calling ShowFunc if set, returning Error if set, or returning nil.
func (llm MockClient) Show(ctx context.Context, model string) error {
	if llm.ShowFunc != nil {
		return llm.ShowFunc(ctx, model)
	}

	if llm.Error != nil {
		return llm.Error
	}

	return nil
}
