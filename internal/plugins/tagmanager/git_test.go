package tagmanager

import (
	"os/exec"
	"testing"
)

func TestCreateAnnotatedTag(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			// Verify correct arguments
			if name != "git" {
				t.Errorf("expected git command, got %s", name)
			}
			if len(args) < 4 || args[0] != "tag" || args[1] != "-a" || args[2] != "v1.0.0" {
				t.Errorf("unexpected args: %v", args)
			}
			return exec.Command("true")
		}

		err := createAnnotatedTag("v1.0.0", "Release 1.0.0")
		if err != nil {
			t.Errorf("createAnnotatedTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'tag already exists' >&2 && exit 1")
		}

		err := createAnnotatedTag("v1.0.0", "Release 1.0.0")
		if err == nil {
			t.Error("createAnnotatedTag() expected error")
		}
		if err.Error() == "" {
			t.Error("expected error message")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		err := createAnnotatedTag("v1.0.0", "Release 1.0.0")
		if err == nil {
			t.Error("createAnnotatedTag() expected error")
		}
	})
}

func TestCreateLightweightTag(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			if name != "git" || len(args) < 2 || args[0] != "tag" {
				t.Errorf("unexpected command: %s %v", name, args)
			}
			return exec.Command("true")
		}

		err := createLightweightTag("v1.0.0")
		if err != nil {
			t.Errorf("createLightweightTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'error' >&2 && exit 1")
		}

		err := createLightweightTag("v1.0.0")
		if err == nil {
			t.Error("createLightweightTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		err := createLightweightTag("v1.0.0")
		if err == nil {
			t.Error("createLightweightTag() expected error")
		}
	})
}

func TestTagExists(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("tag exists", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "v1.0.0")
		}

		exists, err := tagExists("v1.0.0")
		if err != nil {
			t.Errorf("tagExists() error = %v", err)
		}
		if !exists {
			t.Error("tagExists() expected true")
		}
	})

	t.Run("tag does not exist", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "")
		}

		exists, err := tagExists("v1.0.0")
		if err != nil {
			t.Errorf("tagExists() error = %v", err)
		}
		if exists {
			t.Error("tagExists() expected false")
		}
	})

	t.Run("error", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		_, err := tagExists("v1.0.0")
		if err == nil {
			t.Error("tagExists() expected error")
		}
	})
}

func TestGetLatestTag(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "v1.2.3")
		}

		tag, err := getLatestTag()
		if err != nil {
			t.Errorf("getLatestTag() error = %v", err)
		}
		if tag != "v1.2.3" {
			t.Errorf("getLatestTag() = %q, want %q", tag, "v1.2.3")
		}
	})

	t.Run("empty output", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "")
		}

		_, err := getLatestTag()
		if err == nil {
			t.Error("getLatestTag() expected error for empty output")
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'fatal: No names found' >&2 && exit 128")
		}

		_, err := getLatestTag()
		if err == nil {
			t.Error("getLatestTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		_, err := getLatestTag()
		if err == nil {
			t.Error("getLatestTag() expected error")
		}
	})
}

func TestPushTag(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			if name != "git" || len(args) < 3 || args[0] != "push" || args[1] != "origin" {
				t.Errorf("unexpected command: %s %v", name, args)
			}
			return exec.Command("true")
		}

		err := pushTag("v1.0.0")
		if err != nil {
			t.Errorf("pushTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'remote rejected' >&2 && exit 1")
		}

		err := pushTag("v1.0.0")
		if err == nil {
			t.Error("pushTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		err := pushTag("v1.0.0")
		if err == nil {
			t.Error("pushTag() expected error")
		}
	})
}

func TestListTags(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("list all tags", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("printf", "v1.0.0\nv1.1.0\nv2.0.0")
		}

		tags, err := ListTags("")
		if err != nil {
			t.Errorf("ListTags() error = %v", err)
		}
		if len(tags) != 3 {
			t.Errorf("ListTags() returned %d tags, want 3", len(tags))
		}
	})

	t.Run("list with pattern", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			// Verify pattern is passed
			if len(args) < 3 || args[2] != "v1.*" {
				t.Errorf("expected pattern v1.*, got args: %v", args)
			}
			return exec.Command("printf", "v1.0.0\nv1.1.0")
		}

		tags, err := ListTags("v1.*")
		if err != nil {
			t.Errorf("ListTags() error = %v", err)
		}
		if len(tags) != 2 {
			t.Errorf("ListTags() returned %d tags, want 2", len(tags))
		}
	})

	t.Run("empty result", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "")
		}

		tags, err := ListTags("nonexistent*")
		if err != nil {
			t.Errorf("ListTags() error = %v", err)
		}
		if len(tags) != 0 {
			t.Errorf("ListTags() returned %d tags, want 0", len(tags))
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'git error' >&2 && exit 1")
		}

		_, err := ListTags("")
		if err == nil {
			t.Error("ListTags() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		_, err := ListTags("")
		if err == nil {
			t.Error("ListTags() expected error")
		}
	})
}

func TestDeleteTag(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			if name != "git" || len(args) < 3 || args[0] != "tag" || args[1] != "-d" {
				t.Errorf("unexpected command: %s %v", name, args)
			}
			return exec.Command("true")
		}

		err := DeleteTag("v1.0.0")
		if err != nil {
			t.Errorf("DeleteTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'tag not found' >&2 && exit 1")
		}

		err := DeleteTag("v1.0.0")
		if err == nil {
			t.Error("DeleteTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		err := DeleteTag("v1.0.0")
		if err == nil {
			t.Error("DeleteTag() expected error")
		}
	})
}
