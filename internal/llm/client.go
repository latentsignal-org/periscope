// Package llm provides a minimal Anthropic Messages API client used
// by Periscope V2 for turn summarisation and guidance text generation.
//
// The client speaks HTTPS directly to api.anthropic.com so we avoid a
// third-party SDK dependency. Prompt caching is enabled on the system
// prompt so short per-turn calls can amortise the prompt cost.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultEndpoint is the Anthropic Messages API endpoint.
const DefaultEndpoint = "https://api.anthropic.com/v1/messages"

// DefaultAPIVersion is the Anthropic API version header.
const DefaultAPIVersion = "2023-06-01"

// DefaultSummaryModel is used for per-turn summarisation.
const DefaultSummaryModel = "claude-haiku-4-5-20251001"

// DefaultGenerateModel is used for banner text generation.
const DefaultGenerateModel = "claude-haiku-4-5-20251001"

// MaxRetries is the number of HTTP retries for 429/5xx.
const MaxRetries = 4

// Client is the interface the rest of Periscope talks to. Kept small
// so tests can supply a fake.
type Client interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

// Request is a JSON-shaped request body. The Messages field maps
// onto Anthropic's messages[]; SystemCached becomes system[] with a
// cache_control marker.
type Request struct {
	Model        string
	MaxTokens    int
	Temperature  float64
	SystemCached string    // cached via ephemeral cache_control
	Messages     []Message // ordered conversation
}

// Message is one element in the messages[] array.
type Message struct {
	Role    string // "user" | "assistant"
	Content string
}

// Response carries the first text block of the assistant message.
type Response struct {
	Text  string
	Model string
	Usage Usage
}

// Usage mirrors the usage block returned by the API.
type Usage struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
}

// Config configures an HTTPClient.
type Config struct {
	APIKey     string
	Endpoint   string
	APIVersion string
	HTTPClient *http.Client
}

// HTTPClient is the production Client implementation.
type HTTPClient struct {
	cfg Config
	now func() time.Time
}

// NewHTTPClient constructs a client. If APIKey is empty the returned
// client returns ErrNoAPIKey on every call.
func NewHTTPClient(cfg Config) *HTTPClient {
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultEndpoint
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = DefaultAPIVersion
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &HTTPClient{cfg: cfg, now: time.Now}
}

// NewFromEnv returns a client configured from ANTHROPIC_API_KEY and
// optional ANTHROPIC_BASE_URL. Returns (nil, false) when the key is
// missing, so callers can gate cleanly.
func NewFromEnv() (*HTTPClient, bool) {
	key := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	if key == "" {
		return nil, false
	}
	endpoint := strings.TrimSpace(os.Getenv("ANTHROPIC_BASE_URL"))
	if endpoint != "" && !strings.HasSuffix(endpoint, "/v1/messages") {
		endpoint = strings.TrimRight(endpoint, "/") + "/v1/messages"
	}
	return NewHTTPClient(Config{APIKey: key, Endpoint: endpoint}), true
}

// ErrNoAPIKey indicates that the client is not configured with a key.
var ErrNoAPIKey = errors.New("llm: ANTHROPIC_API_KEY is not set")

// Complete sends a single messages request and returns the first
// text block. Retries on 429 and 5xx with exponential backoff.
func (c *HTTPClient) Complete(
	ctx context.Context, req Request,
) (Response, error) {
	if c == nil || c.cfg.APIKey == "" {
		return Response{}, ErrNoAPIKey
	}
	if req.Model == "" {
		req.Model = DefaultGenerateModel
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 1024
	}

	body, err := buildRequestBody(req)
	if err != nil {
		return Response{}, fmt.Errorf("llm: build body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, backoff(attempt)); err != nil {
				return Response{}, err
			}
		}
		httpReq, err := http.NewRequestWithContext(
			ctx, http.MethodPost, c.cfg.Endpoint,
			bytes.NewReader(body),
		)
		if err != nil {
			return Response{}, fmt.Errorf("llm: build request: %w", err)
		}
		httpReq.Header.Set("content-type", "application/json")
		httpReq.Header.Set("x-api-key", c.cfg.APIKey)
		httpReq.Header.Set("anthropic-version", c.cfg.APIVersion)
		httpReq.Header.Set(
			"anthropic-beta",
			"prompt-caching-2024-07-31",
		)

		resp, err := c.cfg.HTTPClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("llm: request: %w", err)
			if !retryable(0) {
				return Response{}, lastErr
			}
			continue
		}
		parsed, retriable, err := parseResponse(resp)
		resp.Body.Close()
		if err == nil {
			return parsed, nil
		}
		lastErr = err
		if !retriable {
			return Response{}, err
		}
	}
	return Response{}, fmt.Errorf(
		"llm: exhausted retries: %w", lastErr,
	)
}

func buildRequestBody(req Request) ([]byte, error) {
	type cacheCtl struct {
		Type string `json:"type"`
	}
	type sysBlock struct {
		Type         string    `json:"type"`
		Text         string    `json:"text"`
		CacheControl *cacheCtl `json:"cache_control,omitempty"`
	}
	type msgBlock struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	body := map[string]any{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.SystemCached != "" {
		body["system"] = []sysBlock{{
			Type:         "text",
			Text:         req.SystemCached,
			CacheControl: &cacheCtl{Type: "ephemeral"},
		}}
	}
	msgs := make([]msgBlock, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, msgBlock{Role: m.Role, Content: m.Content})
	}
	body["messages"] = msgs
	return json.Marshal(body)
}

func parseResponse(resp *http.Response) (Response, bool, error) {
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, true, fmt.Errorf(
			"llm: read body: %w", err,
		)
	}
	if resp.StatusCode >= 500 ||
		resp.StatusCode == http.StatusTooManyRequests {
		return Response{}, true, fmt.Errorf(
			"llm: http %d: %s",
			resp.StatusCode, truncate(string(payload), 400),
		)
	}
	if resp.StatusCode != http.StatusOK {
		return Response{}, false, fmt.Errorf(
			"llm: http %d: %s",
			resp.StatusCode, truncate(string(payload), 400),
		)
	}
	var parsed struct {
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage Usage `json:"usage"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return Response{}, false, fmt.Errorf(
			"llm: decode: %w body=%s",
			err, truncate(string(payload), 400),
		)
	}
	var text strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return Response{
		Text:  text.String(),
		Model: parsed.Model,
		Usage: parsed.Usage,
	}, false, nil
}

func retryable(status int) bool {
	return status == 0 ||
		status == http.StatusTooManyRequests ||
		status >= 500
}

func backoff(attempt int) time.Duration {
	base := time.Duration(1<<attempt) * 500 * time.Millisecond
	jitter := time.Duration(rand.Int64N(int64(base / 2)))
	return base + jitter
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
