package db

import "strings"

// automatedPrefixes are first_message prefixes that identify
// automated (roborev) sessions. Matched case-sensitively.
// Combined with the single-turn gate (user_message_count <= 1)
// to avoid misclassifying interactive sessions.
var automatedPrefixes = []string{
	"You are a code reviewer.",
	"You are a security code reviewer.",
	"You are a design reviewer.",
	"You are a code assistant. Your task is to address",
	"You are a code review insights analyst.",
	"You are reviewing whether an implementation matches",
	"You are a plan document reviewer.",
	"You are a spec document reviewer.",
	"You are summarizing a day of AI agent activity.",
	"You are analyzing AI agent sessions.",
	"## Analysis Request",
	"# Fix Request",
}

// automatedSubstrings are patterns matched anywhere in the
// first message. Used for catch-all markers embedded in
// longer prompts.
var automatedSubstrings = []string{
	"invoked by roborev to perform this review",
}

// IsAutomatedSession returns true if the first message
// matches a known automated review/fix prompt pattern.
func IsAutomatedSession(firstMessage string) bool {
	for _, prefix := range automatedPrefixes {
		if strings.HasPrefix(firstMessage, prefix) {
			return true
		}
	}
	for _, sub := range automatedSubstrings {
		if strings.Contains(firstMessage, sub) {
			return true
		}
	}
	return false
}
