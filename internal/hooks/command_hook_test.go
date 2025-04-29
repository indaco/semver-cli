package hooks

import (
	"testing"
)

func TestCommandHook_Run_Success(t *testing.T) {
	h := CommandHook{
		Name:    "echo-test",
		Command: "echo 'hello world'",
	}

	if err := h.Run(); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestCommandHook_Run_Failure(t *testing.T) {
	h := CommandHook{
		Name:    "fail-test",
		Command: "exit 1",
	}

	if err := h.Run(); err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestCommandHook_HookName(t *testing.T) {
	h := CommandHook{Name: "hook-name"}
	if got := h.HookName(); got != "hook-name" {
		t.Errorf("expected 'hook-name', got %q", got)
	}
}
