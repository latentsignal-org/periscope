package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeTerminal(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain text",
			in:   "hello world",
			want: "hello world",
		},
		{
			name: "newline and tab preserved; CR stripped",
			in:   "line1\nline2\ttab\rcr",
			want: "line1\nline2\ttabcr",
		},
		{
			// Bare \r on a TTY returns the cursor to column 0
			// and lets later text overwrite earlier output —
			// a spoofing vector even without ANSI escapes.
			name: "strips bare CR to block overwrite attacks",
			in:   "safe output\rEVIL",
			want: "safe outputEVIL",
		},
		{
			// CRLF from Windows-style sources collapses to LF;
			// terminals treat \n alone as a newline so display
			// still looks right.
			name: "CRLF collapses to LF",
			in:   "line1\r\nline2",
			want: "line1\nline2",
		},
		{
			name: "strips ESC (C0)",
			in:   "safe\x1bdangerous",
			want: "safedangerous",
		},
		{
			name: "strips ANSI CSI (cursor up)",
			in:   "before\x1b[1Aoverwritten",
			want: "before[1Aoverwritten",
		},
		{
			// ESC (0x1b) is dropped, so the OSC introducer is
			// broken. The literal "]8;;…" bytes remain but are
			// no longer interpreted as an escape sequence.
			name: "strips OSC 8 hyperlink escape",
			in:   "click \x1b]8;;https://evil.example\x1b\\me\x1b]8;;\x1b\\",
			want: "click ]8;;https://evil.example\\me]8;;\\",
		},
		{
			// BEL (0x07) is a C0 control and also dropped.
			name: "strips OSC 52 clipboard escape",
			in:   "benign \x1b]52;c;ZXZpbA==\x07 end",
			want: "benign ]52;c;ZXZpbA== end",
		},
		{
			name: "strips title-set OSC 2",
			in:   "\x1b]2;rm -rf /\x07after",
			want: "]2;rm -rf /after",
		},
		{
			name: "strips NUL and DEL",
			in:   "a\x00b\x7fc",
			want: "abc",
		},
		{
			name: "strips raw C1 bytes",
			in:   "x\x80y\x9fz",
			want: "xyz",
		},
		{
			name: "preserves multibyte UTF-8",
			in:   "héllo 日本語 😀",
			want: "héllo 日本語 😀",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeTerminal(tc.in)
			assert.Equal(t, tc.want, got,
				"sanitizeTerminal(%q)", tc.in)
			// Post-condition: no ESC byte or NUL in output.
			assert.False(t, strings.ContainsRune(got, 0x1b),
				"output still contains ESC: %q", got)
			assert.False(t, strings.ContainsRune(got, 0),
				"output still contains NUL: %q", got)
		})
	}
}
