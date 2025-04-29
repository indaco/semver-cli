package hooks

import (
	"testing"

	"github.com/indaco/semver-cli/internal/config"
)

func TestLoadPreReleaseHooksFromConfig(t *testing.T) {
	t.Cleanup(func() { ResetPreReleaseHooks() })

	cfg := &config.Config{
		PreReleaseHooks: []map[string]config.PreReleaseHookConfig{
			{
				"run-tests": {Command: "go test ./..."},
			},
			{
				"check-license": {Command: "./scripts/check_license.sh"},
			},
		},
	}

	err := LoadPreReleaseHooksFromConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	hooks := GetPreReleaseHooks()
	if len(hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(hooks))
	}

	if hooks[0].HookName() != "run-tests" {
		t.Errorf("expected first hook 'run-tests', got %q", hooks[0].HookName())
	}
	if hooks[1].HookName() != "check-license" {
		t.Errorf("expected second hook 'check-license', got %q", hooks[1].HookName())
	}
}
