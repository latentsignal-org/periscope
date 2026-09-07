package ssh

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteTarStderrBenign(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   bool
	}{
		{
			name: "file changed with delayed-exit fallout",
			stderr: "tar: home/wes/.claude/x.jsonl: " +
				"file changed as we read it\n" +
				"tar: Exiting with failure status " +
				"due to previous errors",
			want: true,
		},
		{
			name: "file removed before read",
			stderr: "tar: home/wes/.codex/y.json: " +
				"file removed before we read it",
			want: true,
		},
		{
			name: "permission denied is fatal",
			stderr: "tar: home/wes/.claude: Cannot open: " +
				"Permission denied\n" +
				"tar: Exiting with failure status " +
				"due to previous errors",
			want: false,
		},
		{
			name:   "remote truncation is fatal",
			stderr: "tar: Unexpected EOF in archive",
			want:   false,
		},
		{
			name: "delayed-exit summary alone is not benign",
			stderr: "tar: Exiting with failure status " +
				"due to previous errors",
			want: false,
		},
		{
			name:   "empty stderr is not benign",
			stderr: "",
			want:   false,
		},
		{
			name: "ssh connection failure is fatal",
			stderr: "ssh: connect to host devbox port 22: " +
				"Connection refused",
			want: false,
		},
		{
			name: "benign mixed with fatal stays fatal",
			stderr: "tar: a: file changed as we read it\n" +
				"tar: b: Cannot stat: No such file or directory",
			want: false,
		},
		{
			name: "benign phrase in path with real error stays fatal",
			stderr: "tar: home/wes/file changed as we read it: " +
				"Cannot open: Permission denied",
			want: false,
		},
		{
			name: "genuine warning on a path containing the phrase",
			stderr: "tar: home/file changed as we read it: " +
				"file changed as we read it",
			want: true,
		},
		{
			name: "bsd delayed-exit summary with trailing period",
			stderr: "tar: x.jsonl: file changed as we read it\n" +
				"tar: Error exit delayed from previous errors.",
			want: true,
		},
		{
			// GNU tar uses a capital F here (create.c), unlike the
			// lowercase "file changed as we read it".
			name: "capitalized GNU file-removed warning is benign",
			stderr: "tar: home/wes/.codex/x.json: " +
				"File removed before we read it",
			want: true,
		},
		{
			name: "capitalized file-changed warning is benign",
			stderr: "tar: home/wes/.claude/y.jsonl: " +
				"File changed as we read it",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &commandError{
				Host:   "devbox",
				Stderr: tt.stderr,
				Err:    errors.New("exit status 1"),
			}
			assert.Equal(t, tt.want, remoteTarStderrBenign(err))
		})
	}
}

func TestRemoteTarStderrBenignNonCommandError(t *testing.T) {
	// Anything that is not a classified SSH command failure must
	// never be treated as benign.
	assert.False(t, remoteTarStderrBenign(errors.New("boom")))
	assert.False(t, remoteTarStderrBenign(nil))
}

func TestRemoteTarStderrBenignWrapped(t *testing.T) {
	// The classifier must see through fmt.Errorf wrapping, since
	// downloadAndExtract wraps cleanup errors as "ssh tar: %w".
	base := &commandError{
		Host:   "devbox",
		Stderr: "tar: f: file changed as we read it",
		Err:    errors.New("exit status 1"),
	}
	wrapped := fmt.Errorf("ssh tar: %w", base)
	assert.True(t, remoteTarStderrBenign(wrapped))
}
