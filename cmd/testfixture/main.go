package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wesm/agentsview/internal/db"
)

type sessionSpec struct {
	project          string
	suffix           string
	msgCount         int
	userMsgCount     int
	parentSessionID  string
	relationshipType string
}

var specs = []sessionSpec{
	{"project-alpha", "small-2", 2, 2, "", ""},
	{"project-alpha", "small-5", 5, 3, "", ""},
	{"project-beta", "mixed-content-7", 7, 3, "", ""},
	{"project-beta", "medium-8", 8, 4, "", ""},
	{"project-beta", "medium-100", 100, 50, "", ""},
	{"project-gamma", "large-200", 200, 100, "", ""},
	{"project-gamma", "large-1500", 1500, 750, "", ""},
	{"project-delta", "xlarge-5500", 5500, 2750, "", ""},

	// Sub-agent and fork sessions: must NOT appear in session
	// list, stats, or analytics summary counts.
	{"project-alpha", "subagent-1", 12, 6,
		"test-session-small-5", "subagent"},
	{"project-alpha", "subagent-2", 8, 4,
		"test-session-small-5", "subagent"},
	{"project-beta", "fork-1", 15, 7,
		"test-session-medium-8", "fork"},

	// Empty session (0 messages): must also be excluded.
	{"project-gamma", "empty-0", 0, 0, "", ""},
}

func main() {
	out := flag.String("out", "", "output database path")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, "usage: testfixture -out <path>")
		os.Exit(1)
	}

	if err := os.Remove(*out); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		log.Fatalf("removing existing db: %v", err)
	}

	database, err := db.Open(*out)
	if err != nil {
		log.Fatalf("opening db: %v", err)
	}
	defer database.Close()

	// Seed model pricing for usage page e2e tests.
	if err := database.UpsertModelPricing([]db.ModelPricing{
		{
			ModelPattern:         "claude-sonnet-4-20250514",
			InputPerMTok:         3.0,
			OutputPerMTok:        15.0,
			CacheCreationPerMTok: 3.75,
			CacheReadPerMTok:     0.30,
		},
		{
			ModelPattern:         "claude-opus-4-20250514",
			InputPerMTok:         15.0,
			OutputPerMTok:        75.0,
			CacheCreationPerMTok: 18.75,
			CacheReadPerMTok:     1.50,
		},
	}); err != nil {
		log.Fatalf("seeding model pricing: %v", err)
	}

	// Use a recent base date so fixture data stays within the
	// default 1-year analytics window.
	base := time.Now().UTC().AddDate(0, 0, -30).
		Truncate(24 * time.Hour).Add(10 * time.Hour)

	for i, spec := range specs {
		if err := createSessionFixture(
			database, spec, i, base,
		); err != nil {
			log.Fatalf("creating fixture %s: %v", spec.suffix, err)
		}
		fmt.Printf(
			"  test-session-%s: %d messages\n",
			spec.suffix, spec.msgCount,
		)
	}

	fmt.Printf("Fixture DB written to %s\n", *out)
}

func ptr[T any](v T) *T { return &v }

func createSessionFixture(
	database *db.DB, spec sessionSpec,
	index int, base time.Time,
) error {
	sessionID := fmt.Sprintf("test-session-%s", spec.suffix)
	startedAt := base.Add(
		time.Duration(index) * 24 * time.Hour,
	)
	endedAt := startedAt.Add(
		time.Duration(spec.msgCount) * time.Minute,
	)

	sess := db.Session{
		ID:               sessionID,
		Project:          spec.project,
		Machine:          "test-machine",
		Agent:            "claude",
		StartedAt:        ptr(startedAt.Format(time.RFC3339Nano)),
		EndedAt:          ptr(endedAt.Format(time.RFC3339Nano)),
		MessageCount:     spec.msgCount,
		UserMessageCount: spec.userMsgCount,
		RelationshipType: spec.relationshipType,
	}
	if spec.parentSessionID != "" {
		sess.ParentSessionID = ptr(spec.parentSessionID)
	}
	if spec.msgCount > 0 {
		sess.FirstMessage = ptr(
			fmt.Sprintf("First message for %s", spec.project),
		)
	}
	if err := database.UpsertSession(sess); err != nil {
		return fmt.Errorf("upserting session: %w", err)
	}

	if spec.msgCount == 0 {
		return nil
	}

	model := "claude-sonnet-4-20250514"
	if index%3 == 1 {
		model = "claude-opus-4-20250514"
	}

	var msgs []db.Message
	if spec.suffix == "mixed-content-7" {
		msgs = generateMixedContentMessages(
			sessionID, startedAt, model,
		)
	} else {
		msgs = generateMessages(
			sessionID, spec.msgCount, startedAt, model,
		)
	}
	if err := database.InsertMessages(msgs); err != nil {
		return fmt.Errorf("inserting messages: %w", err)
	}
	return nil
}

func generateMessages(
	sessionID string, count int,
	start time.Time, model string,
) []db.Message {
	msgs := make([]db.Message, 0, count)
	for i := range count {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}

		ts := start.Add(time.Duration(i) * time.Minute)
		content := generateContent(role, i, count)

		msg := db.Message{
			SessionID:     sessionID,
			Ordinal:       i,
			Role:          role,
			Content:       content,
			Timestamp:     ts.Format(time.RFC3339Nano),
			HasThinking:   role == "assistant" && i%5 == 0,
			HasToolUse:    role == "assistant" && i%3 == 0,
			ContentLength: len(content),
		}

		if role == "assistant" && model != "" {
			msg.Model = model
			inputTok := 500 + (i*137)%2000
			outputTok := 200 + (i*89)%800
			cacheCr := 50 + (i*31)%200
			cacheRd := 1000 + (i*53)%4000
			msg.TokenUsage = json.RawMessage(
				fmt.Sprintf(
					`{"input_tokens":%d,`+
						`"output_tokens":%d,`+
						`"cache_creation_input_tokens":%d,`+
						`"cache_read_input_tokens":%d}`,
					inputTok, outputTok,
					cacheCr, cacheRd,
				),
			)
		}

		msgs = append(msgs, msg)
	}
	return msgs
}

func generateMixedContentMessages(
	sessionID string, start time.Time, model string,
) []db.Message {
	type spec struct {
		role        string
		content     string
		hasThinking bool
		hasToolUse  bool
	}

	specs := []spec{
		{
			role:    "user",
			content: "Help me read a file",
		},
		{
			role: "assistant",
			content: "[Thinking]\nLet me analyze..." +
				"\n\nHere is my analysis.",
			hasThinking: true,
		},
		{
			role:    "user",
			content: "Now check the directory",
		},
		{
			role:       "assistant",
			content:    "[Read /src/main.ts]\nconst app = express();",
			hasToolUse: true,
		},
		{
			role:       "assistant",
			content:    "[Bash]\nls -la /src",
			hasToolUse: true,
		},
		{
			role: "assistant",
			content: "[Thinking]\nGemini-style reasoning\n" +
				"[/Thinking]\n\n" +
				"This is the visible response after thinking.",
			hasThinking: true,
		},
		{
			role:    "user",
			content: "Thanks",
		},
	}

	msgs := make([]db.Message, 0, len(specs))
	for i, s := range specs {
		ts := start.Add(time.Duration(i) * time.Minute)
		msg := db.Message{
			SessionID:     sessionID,
			Ordinal:       i,
			Role:          s.role,
			Content:       s.content,
			Timestamp:     ts.Format(time.RFC3339Nano),
			HasThinking:   s.hasThinking,
			HasToolUse:    s.hasToolUse,
			ContentLength: len(s.content),
		}
		if s.role == "assistant" && model != "" {
			msg.Model = model
			inputTok := 400 + (i*113)%1500
			outputTok := 150 + (i*67)%600
			cacheCr := 30 + (i*23)%150
			cacheRd := 800 + (i*41)%3000
			msg.TokenUsage = json.RawMessage(
				fmt.Sprintf(
					`{"input_tokens":%d,`+
						`"output_tokens":%d,`+
						`"cache_creation_input_tokens":%d,`+
						`"cache_read_input_tokens":%d}`,
					inputTok, outputTok,
					cacheCr, cacheRd,
				),
			)
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

func generateContent(role string, idx, total int) string {
	if role == "user" {
		return fmt.Sprintf(
			"User message %d of %d. "+
				"Please help me with this task. "+
				"I need to understand how the code works.",
			idx, total,
		)
	}
	return fmt.Sprintf(
		"Assistant response %d of %d. "+
			"Here is my analysis of the code. "+
			"The implementation follows standard patterns "+
			"and uses well-known libraries. "+
			"Let me explain the key components.",
		idx, total,
	)
}
