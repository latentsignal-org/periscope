package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCompleteSuccess verifies happy-path request/response plumbing:
// headers, cache-control on system, and extraction of the first text
// block.
func TestCompleteSuccess(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter, r *http.Request,
	) {
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Errorf("x-api-key = %q", got)
		}
		if got := r.Header.Get("anthropic-version"); got == "" {
			t.Errorf("anthropic-version missing")
		}
		if got := r.Header.Get("anthropic-beta"); !strings.Contains(
			got, "prompt-caching",
		) {
			t.Errorf("anthropic-beta = %q, want prompt-caching", got)
		}
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{
			"model": "claude-haiku-4-5",
			"content": [{"type":"text","text":"hello"}],
			"usage": {"input_tokens":1,"output_tokens":2}
		}`))
	}))
	defer srv.Close()

	c := NewHTTPClient(Config{
		APIKey:   "test-key",
		Endpoint: srv.URL,
	})
	resp, err := c.Complete(context.Background(), Request{
		Model:        "claude-haiku-4-5",
		MaxTokens:    100,
		SystemCached: "system text",
		Messages:     []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Text != "hello" {
		t.Errorf("Text = %q, want hello", resp.Text)
	}

	sys, ok := gotBody["system"].([]any)
	if !ok || len(sys) == 0 {
		t.Fatalf("system block missing: %v", gotBody["system"])
	}
	first := sys[0].(map[string]any)
	cc, ok := first["cache_control"].(map[string]any)
	if !ok || cc["type"] != "ephemeral" {
		t.Errorf("cache_control = %v, want ephemeral", first["cache_control"])
	}
}

// TestCompleteNoKey verifies a nil-key client short-circuits to
// ErrNoAPIKey without any network call.
func TestCompleteNoKey(t *testing.T) {
	c := NewHTTPClient(Config{})
	_, err := c.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "x"}},
	})
	if err == nil {
		t.Fatalf("expected ErrNoAPIKey")
	}
}
