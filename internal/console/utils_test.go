package console

import (
	"fmt"
	"testing"

	"github.com/indaco/verso/internal/testutils"
)

func TestPrintSuccess(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
		msg     string
		wantOut string
	}{
		{
			name:    "with color",
			noColor: false,
			msg:     "Success!",
			wantOut: fmt.Sprintf("%s%s%s", colorGreen, "Success!", colorReset),
		},
		{
			name:    "without color",
			noColor: true,
			msg:     "Success!",
			wantOut: "Success!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := testutils.CaptureStdout(func() {
				SetNoColor(tt.noColor)
				PrintSuccess(tt.msg)
			})
			if err != nil {
				t.Fatalf("CaptureStdout() error: %v", err)
			}
			if gotOut != tt.wantOut {
				t.Errorf("PrintSuccess() output = %q, want %q", gotOut, tt.wantOut)
			}
		})
	}
}

func TestPrintFailure(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
		msg     string
		wantOut string
	}{
		{
			name:    "with color",
			noColor: false,
			msg:     "Failure!",
			wantOut: fmt.Sprintf("%s%s%s", colorRed, "Failure!", colorReset),
		},
		{
			name:    "without color",
			noColor: true,
			msg:     "Failure!",
			wantOut: "Failure!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := testutils.CaptureStdout(func() {
				SetNoColor(tt.noColor)
				PrintFailure(tt.msg)
			})
			if err != nil {
				t.Fatalf("CaptureStdout() error: %v", err)
			}
			if gotOut != tt.wantOut {
				t.Errorf("PrintFailure() output = %q, want %q", gotOut, tt.wantOut)
			}
		})
	}
}
