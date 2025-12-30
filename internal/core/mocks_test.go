package core

import (
	"context"
	"errors"
	"io/fs"
	"testing"
)

func TestMockFileSystem(t *testing.T) {
	mockFS := NewMockFileSystem()

	t.Run("write and read file", func(t *testing.T) {
		content := []byte("test content")
		err := mockFS.WriteFile("/test/file.txt", content, 0644)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		data, err := mockFS.ReadFile("/test/file.txt")
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}

		if string(data) != string(content) {
			t.Errorf("expected %q, got %q", string(content), string(data))
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := mockFS.ReadFile("/nonexistent")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected fs.ErrNotExist, got %v", err)
		}
	})

	t.Run("stat file", func(t *testing.T) {
		mockFS.SetFile("/stat/test.txt", []byte("hello"))
		info, err := mockFS.Stat("/stat/test.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if info.Size() != 5 {
			t.Errorf("expected size 5, got %d", info.Size())
		}
		if info.IsDir() {
			t.Error("expected file, got directory")
		}
	})

	t.Run("mkdir and stat directory", func(t *testing.T) {
		err := mockFS.MkdirAll("/test/dir", 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		info, err := mockFS.Stat("/test/dir")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory, got file")
		}
	})

	t.Run("remove file", func(t *testing.T) {
		mockFS.SetFile("/remove/test.txt", []byte("to be removed"))
		err := mockFS.Remove("/remove/test.txt")
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		_, err = mockFS.ReadFile("/remove/test.txt")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected fs.ErrNotExist after removal, got %v", err)
		}
	})

	t.Run("error injection", func(t *testing.T) {
		mockFS.ReadErr = errors.New("read error")
		_, err := mockFS.ReadFile("/any/path")
		if err == nil || err.Error() != "read error" {
			t.Errorf("expected read error, got %v", err)
		}
		mockFS.ReadErr = nil
	})
}

func TestMockCommandExecutor(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	ctx := context.Background()

	t.Run("run with default success", func(t *testing.T) {
		err := mockExec.Run(ctx, ".", "echo", "hello")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mockExec.Calls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mockExec.Calls))
		}
		if mockExec.Calls[0].Command != "echo" {
			t.Errorf("expected 'echo' command, got %q", mockExec.Calls[0].Command)
		}
	})

	t.Run("output with set response", func(t *testing.T) {
		mockExec.SetResponse("git describe --tags", "v1.2.3\n")
		output, err := mockExec.Output(ctx, ".", "git", "describe", "--tags")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output != "v1.2.3\n" {
			t.Errorf("expected 'v1.2.3\\n', got %q", output)
		}
	})

	t.Run("error injection", func(t *testing.T) {
		mockExec.SetError("make test", errors.New("test failed"))
		_, err := mockExec.Output(ctx, ".", "make", "test")
		if err == nil || err.Error() != "test failed" {
			t.Errorf("expected 'test failed' error, got %v", err)
		}
	})

	t.Run("default error", func(t *testing.T) {
		mockExec := NewMockCommandExecutor()
		mockExec.DefaultError = errors.New("default error")

		err := mockExec.Run(ctx, ".", "unknown", "command")
		if err == nil || err.Error() != "default error" {
			t.Errorf("expected 'default error', got %v", err)
		}
	})
}

func TestMockGitClient(t *testing.T) {
	mockGit := NewMockGitClient()
	ctx := context.Background()

	t.Run("describe tags", func(t *testing.T) {
		mockGit.TagOutput = "v1.0.0"
		tag, err := mockGit.DescribeTags(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tag != "v1.0.0" {
			t.Errorf("expected 'v1.0.0', got %q", tag)
		}
	})

	t.Run("describe tags error", func(t *testing.T) {
		mockGit.TagError = errors.New("no tags")
		_, err := mockGit.DescribeTags(ctx)
		if err == nil || err.Error() != "no tags" {
			t.Errorf("expected 'no tags' error, got %v", err)
		}
		mockGit.TagError = nil
	})

	t.Run("clone", func(t *testing.T) {
		err := mockGit.Clone(ctx, "https://example.com/repo.git", "/tmp/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !mockGit.IsValidRepo("/tmp/repo") {
			t.Error("expected cloned repo to be marked as valid")
		}
	})

	t.Run("is valid repo", func(t *testing.T) {
		mockGit.IsValidRepos["/existing/repo"] = true

		if !mockGit.IsValidRepo("/existing/repo") {
			t.Error("expected /existing/repo to be valid")
		}
		if mockGit.IsValidRepo("/nonexistent") {
			t.Error("expected /nonexistent to be invalid")
		}
	})
}

func TestOSFileSystem(t *testing.T) {
	osFS := NewOSFileSystem()
	tmpDir := t.TempDir()

	t.Run("write and read file", func(t *testing.T) {
		path := tmpDir + "/test.txt"
		content := []byte("hello world")

		err := osFS.WriteFile(path, content, 0600)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		data, err := osFS.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}

		if string(data) != string(content) {
			t.Errorf("expected %q, got %q", string(content), string(data))
		}
	})

	t.Run("stat", func(t *testing.T) {
		info, err := osFS.Stat(tmpDir)
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory")
		}
	})

	t.Run("mkdir all", func(t *testing.T) {
		path := tmpDir + "/a/b/c"
		err := osFS.MkdirAll(path, 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		info, err := osFS.Stat(path)
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory")
		}
	})
}
