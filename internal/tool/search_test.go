package tool

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchExecute(t *testing.T) {
	tests := []struct {
		name           string
		args           string
		mockStatusCode int
		mockResponse   string
		wantContains   string
		wantErr        bool
		err            error
	}{
		{
			name:           "successful search",
			args:           `{"query": "test query"}`,
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"results":[{"title":"Test Title","url":"https://test.com","content":"Test content"}]}`,
			wantContains:   "Test Title",
		},
		{
			name:           "multiple results",
			args:           `{"query": "test query"}`,
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"results":[{"title":"First","url":"https://first.com","content":"First content"},{"title":"Second","url":"https://second.com","content":"Second content"}]}`,
			wantContains:   "Second",
		},
		{
			name:           "empty results",
			args:           `{"query": "test query"}`,
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"results":[]}`,
			wantContains:   "",
		},
		{
			name:           "http error",
			args:           `{"query": "test query"}`,
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   `{"error": "internal error"}`,
			wantErr:        true,
			err:            ErrSearchFailed,
		},
		{
			name:    "invalid json args",
			args:    `{"query": }`,
			wantErr: true,
			err:     ErrParseArgs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			search := Search{
				APIKey:     "test-key",
				MaxResults: 5,
				URL:        server.URL,
			}

			got, err := search.Execute(context.Background(), json.RawMessage(tt.args))

			if tt.wantErr {
				if err == nil {
					t.Fatal("Execute() err = nil, want error")
				}

				if !errors.Is(err, tt.err) {
					t.Errorf("Execute() err = %v, want %v", err, tt.err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Execute() err = %v, want nil", err)
			}

			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("Execute() result = %q, want to contain %q", got, tt.wantContains)
			}
		})
	}
}

func TestSearchDefinition(t *testing.T) {
	search := NewSearch("test-key", 5)

	def := search.Definition()

	if def.Type != "function" {
		t.Errorf("Definition() type = %s, want function", def.Type)
	}

	if def.Function.Name != "web_search" {
		t.Errorf("Definition() name = %s, want web_search", def.Function.Name)
	}

	if len(def.Function.Parameters.Required) != 1 || def.Function.Parameters.Required[0] != "query" {
		t.Errorf("Definition() required = %v, want [query]", def.Function.Parameters.Required)
	}
}
