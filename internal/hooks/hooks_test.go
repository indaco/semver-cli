package hooks

import (
	"errors"
	"testing"

	"github.com/indaco/semver-cli/internal/testutils"
)

func TestRunPreReleaseHooks(t *testing.T) {
	tests := []struct {
		name       string
		skip       bool
		hooks      []PreReleaseHook
		wantErr    bool
		errMessage string
	}{
		{
			name:    "skip true",
			skip:    true,
			hooks:   []PreReleaseHook{testutils.MockHook{Name: "hook1"}},
			wantErr: false,
		},
		{
			name:    "no hooks",
			skip:    false,
			hooks:   nil,
			wantErr: false,
		},
		{
			name: "all hooks succeed",
			skip: false,
			hooks: []PreReleaseHook{
				testutils.MockHook{Name: "hook1"},
				testutils.MockHook{Name: "hook2"},
			},
			wantErr: false,
		},
		{
			name: "hook fails",
			skip: false,
			hooks: []PreReleaseHook{
				testutils.MockHook{Name: "hook1"},
				testutils.MockHook{Name: "bad-hook", ShouldErr: true},
				testutils.MockHook{Name: "hook2"}, // Should not be called after failure
			},
			wantErr:    true,
			errMessage: `pre-release hook "bad-hook" failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetPreReleaseHooks()
			t.Cleanup(func() { ResetPreReleaseHooks() })
			for _, h := range tt.hooks {
				RegisterPreReleaseHook(h)
			}

			err := runPreReleaseHooks(tt.skip)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tt.wantErr, err != nil)
			}

			if tt.wantErr && !errors.Is(err, tt.hooks[1].Run()) {
				// error wrapping check
				if err == nil || !containsError(err.Error(), tt.errMessage) {
					t.Errorf("expected error message to contain %q, got: %v", tt.errMessage, err)
				}
			}
		})
	}
}

func containsError(got, want string) bool {
	return got != "" && want != "" && (got == want || (len(got) > len(want) && got[:len(want)] == want) || (len(want) > len(got) && want[:len(got)] == got))
}
