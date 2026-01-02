package testutils

import (
	"testing"
)

func TestMockPlugin(t *testing.T) {
	mock := MockPlugin{
		NameValue:        "test-plugin",
		VersionValue:     "v1.0.0",
		DescriptionValue: "A test plugin",
	}

	if got := mock.Name(); got != "test-plugin" {
		t.Errorf("Name() = %q, want %q", got, "test-plugin")
	}

	if got := mock.Version(); got != "v1.0.0" {
		t.Errorf("Version() = %q, want %q", got, "v1.0.0")
	}

	if got := mock.Description(); got != "A test plugin" {
		t.Errorf("Description() = %q, want %q", got, "A test plugin")
	}
}

func TestMockExtension(t *testing.T) {
	mock := MockExtension{
		NameValue:        "test-extension",
		VersionValue:     "v2.0.0",
		DescriptionValue: "A test extension",
		HooksValue:       []string{"pre-bump", "post-bump"},
		EntryValue:       "run.sh",
	}

	if got := mock.Name(); got != "test-extension" {
		t.Errorf("Name() = %q, want %q", got, "test-extension")
	}

	if got := mock.Version(); got != "v2.0.0" {
		t.Errorf("Version() = %q, want %q", got, "v2.0.0")
	}

	if got := mock.Description(); got != "A test extension" {
		t.Errorf("Description() = %q, want %q", got, "A test extension")
	}

	hooks := mock.Hooks()
	if len(hooks) != 2 {
		t.Errorf("Hooks() length = %d, want 2", len(hooks))
	}
	if hooks[0] != "pre-bump" {
		t.Errorf("Hooks()[0] = %q, want %q", hooks[0], "pre-bump")
	}
	if hooks[1] != "post-bump" {
		t.Errorf("Hooks()[1] = %q, want %q", hooks[1], "post-bump")
	}

	if got := mock.Entry(); got != "run.sh" {
		t.Errorf("Entry() = %q, want %q", got, "run.sh")
	}
}

func TestMockCommitParser(t *testing.T) {
	t.Run("successful parse", func(t *testing.T) {
		mock := MockCommitParser{
			Label: "minor",
			Err:   nil,
		}

		if got := mock.Name(); got != "mock" {
			t.Errorf("Name() = %q, want %q", got, "mock")
		}

		label, err := mock.Parse([]string{"feat: new feature"})
		if err != nil {
			t.Errorf("Parse() unexpected error: %v", err)
		}
		if label != "minor" {
			t.Errorf("Parse() = %q, want %q", label, "minor")
		}
	})

	t.Run("parse with error", func(t *testing.T) {
		mock := MockCommitParser{
			Label: "",
			Err:   ErrMockParseFailed,
		}

		_, err := mock.Parse([]string{})
		if err != ErrMockParseFailed {
			t.Errorf("Parse() error = %v, want %v", err, ErrMockParseFailed)
		}
	})
}

// ErrMockParseFailed is a sentinel error for testing
var ErrMockParseFailed = testError{"mock parse failed"}

type testError struct {
	msg string
}

func (e testError) Error() string { return e.msg }

func TestMockHook(t *testing.T) {
	t.Run("successful hook", func(t *testing.T) {
		mock := MockHook{
			Name:      "pre-release",
			ShouldErr: false,
		}

		if got := mock.HookName(); got != "pre-release" {
			t.Errorf("HookName() = %q, want %q", got, "pre-release")
		}

		if err := mock.Run(); err != nil {
			t.Errorf("Run() unexpected error: %v", err)
		}
	})

	t.Run("failing hook", func(t *testing.T) {
		mock := MockHook{
			Name:      "validation",
			ShouldErr: true,
		}

		if got := mock.HookName(); got != "validation" {
			t.Errorf("HookName() = %q, want %q", got, "validation")
		}

		err := mock.Run()
		if err == nil {
			t.Error("Run() expected error, got nil")
		}
		if err.Error() != "validation failed" {
			t.Errorf("Run() error = %q, want %q", err.Error(), "validation failed")
		}
	})
}

func TestWithMock(t *testing.T) {
	setupCalled := false
	testFuncCalled := false

	WithMock(
		func() {
			setupCalled = true
		},
		func() {
			testFuncCalled = true
		},
	)

	if !setupCalled {
		t.Error("WithMock should call setup function")
	}
	if !testFuncCalled {
		t.Error("WithMock should call test function")
	}
}
