package browse

import (
	"errors"
	"testing"
)

func TestValidateScreenshotPath(t *testing.T) {
	tests := []struct {
		path string
		ok   bool
	}{
		{"/trace/01-filled.png", true},
		{"/trace/02-confirmed.png", true},
		{"/trace/sub/x.png", true},
		{"/trace/./x.png", true},
		{"/handoff/x.png", false},
		{"trace/x.png", false},
		{"/trace/../etc/x.png", false},
		{"/tracex/a.png", false},
		{"../x.png", false},
		{"/trace", false},
		{"/trace/", false},
		{"", false},
	}
	for _, tt := range tests {
		err := ValidateScreenshotPath(tt.path)
		if tt.ok && err != nil {
			t.Errorf("ValidateScreenshotPath(%q) = %v, want nil", tt.path, err)
		}
		if !tt.ok && !errors.Is(err, ErrScreenshotOutsideTrace) {
			t.Errorf("ValidateScreenshotPath(%q) = %v, want ErrScreenshotOutsideTrace", tt.path, err)
		}
	}
}

func TestValidateArgsScreenshot(t *testing.T) {
	bad := ErrScreenshotOutsideTrace
	tests := []struct {
		name string
		args []string
		want error
	}{
		{"default path", []string{"screenshot"}, nil},
		{"flag only", []string{"screenshot", "--full"}, nil},
		{"trace", []string{"screenshot", "/trace/01-filled.png"}, nil},
		{"trace sub", []string{"screenshot", "/trace/sub/x.png"}, nil},
		{"flags around", []string{"--json", "screenshot", "--full", "/trace/a.png", "--annotate"}, nil},
		{"value flags", []string{"screenshot", "--screenshot-format", "jpeg",
			"--screenshot-quality", "80", "--threshold", "0.01", "/trace/a.jpg"}, nil},
		{"dir trace", []string{"screenshot", "--screenshot-dir", "/trace"}, nil},
		{"dir trace sub", []string{"screenshot", "--screenshot-dir=/trace/shots"}, nil},
		{"handoff", []string{"screenshot", "/handoff/x.png"}, bad},
		{"relative", []string{"screenshot", "trace/x.png"}, bad},
		{"traversal", []string{"screenshot", "/trace/../etc/x.png"}, bad},
		{"prefix", []string{"screenshot", "/tracex/a.png"}, bad},
		{"dotdot", []string{"screenshot", "../x.png"}, bad},
		{"second bad", []string{"screenshot", "/trace/a.png", "/tmp/b.png"}, bad},
		{"unknown flag value", []string{"screenshot", "--quality", "80"}, bad},
		{"dir outside", []string{"screenshot", "--screenshot-dir", "./shots"}, bad},
		{"dir eq outside", []string{"screenshot", "--screenshot-dir=/tracex"}, bad},
		{"dir missing", []string{"screenshot", "--screenshot-dir"}, bad},
		{"upload from trace", []string{"upload", "@e1", "/trace/x.pdf"}, ErrUploadOutsideHandoff},
		{"upload bad then screenshot", []string{"upload", "@e1", "/trace/x.pdf", "screenshot"},
			ErrUploadOutsideHandoff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArgs(tt.args)
			if tt.want == nil && err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestValidateUploadPathRefusesTrace(t *testing.T) {
	if err := ValidateUploadPath("/trace/x.pdf"); !errors.Is(err, ErrUploadOutsideHandoff) {
		t.Fatalf("ValidateUploadPath(/trace/x.pdf) = %v, want ErrUploadOutsideHandoff", err)
	}
}
