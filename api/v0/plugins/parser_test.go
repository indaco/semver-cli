package plugins

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

/* ------------------------------------------------------------------------- */
/* MOCKS                                                                     */
/* ------------------------------------------------------------------------- */

type fakeCommitParser struct {
	name  string
	parse func([]string) (string, error)
}

func (f fakeCommitParser) Name() string {
	return f.name
}

func (f fakeCommitParser) Parse(commits []string) (string, error) {
	return f.parse(commits)
}

/* ------------------------------------------------------------------------- */
/* TESTS                                                                     */
/* ------------------------------------------------------------------------- */

func TestRegisterCommitParser_SetsOnce(t *testing.T) {
	ResetCommitParser()
	defer ResetCommitParser()

	p := fakeCommitParser{name: "first"}
	RegisterCommitParser(p)

	registered := GetCommitParser()
	if registered == nil || registered.Name() != "first" {
		t.Fatalf("expected first parser to be registered, got %v", registered)
	}

	RegisterCommitParser(fakeCommitParser{name: "second"})
	still := GetCommitParser()
	if still.Name() != "first" {
		t.Errorf("expected original parser to remain, got %q", still.Name())
	}
}

func TestCommitParser_ParseDelegates(t *testing.T) {
	ResetCommitParser()
	defer ResetCommitParser()

	expected := "minor"
	RegisterCommitParser(fakeCommitParser{
		name: "test",
		parse: func(commits []string) (string, error) {
			return expected, nil
		},
	})

	parser := GetCommitParser()
	result, err := parser.Parse([]string{"feat: add button"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCommitParser_ParseReturnsError(t *testing.T) {
	ResetCommitParser()
	defer ResetCommitParser()

	RegisterCommitParser(fakeCommitParser{
		name: "failing",
		parse: func(commits []string) (string, error) {
			return "", errors.New("parse failed")
		},
	})

	_, err := GetCommitParser().Parse([]string{"invalid commit"})
	if err == nil {
		t.Errorf("expected error from parser, got nil")
	}
}

func TestGetCommitParser_ReturnsNilIfUnset(t *testing.T) {
	ResetCommitParser()
	defer ResetCommitParser()

	if GetCommitParser() != nil {
		t.Errorf("expected nil parser when unset")
	}
}

func TestRegisterCommitParser_Single(t *testing.T) {
	defer ResetCommitParser()

	p := &fakeCommitParser{name: "primary"}
	RegisterCommitParser(p)

	if GetCommitParser() == nil || GetCommitParser().Name() != "primary" {
		t.Errorf("expected commit parser to be 'primary'")
	}
}

func TestRegisterCommitParser_SecondShowsWarning(t *testing.T) {
	ResetCommitParser()
	defer ResetCommitParser()

	// Capture stderr
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Register the first parser
	RegisterCommitParser(&fakeCommitParser{name: "first"})

	// Register a second parser, should trigger warning
	RegisterCommitParser(&fakeCommitParser{name: "second"})

	// Restore stderr and read output
	_ = w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	expected := "Ignoring commit parser \"second\": another parser (\"first\") is already registered."
	if !strings.Contains(output, expected) {
		t.Errorf("expected warning %q, got: %s", expected, output)
	}
}
