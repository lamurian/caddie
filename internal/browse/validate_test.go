package browse

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateSession(t *testing.T) {
	label63 := strings.Repeat("a", 63)
	tests := []struct {
		name string
		ok   bool
	}{
		{"careers.who.int", true},
		{"inspira-un", true},
		{"a", true},
		{"x1.y2", true},
		{label63, true},
		{strings.Repeat(label63+".", 3) + strings.Repeat("a", 61), true}, // 253
		{"", false},
		{"../x", false},
		{"a b", false},
		{"a/b", false},
		{"/", false},
		{".", false},
		{".a", false},
		{"a.", false},
		{"a..b", false},
		{"-a", false},
		{"a-", false},
		{"a.-b", false},
		{"Careers.WHO.int", false},
		{"a_b", false},
		{"é.org", false},
		{label63 + "a", false},
		{strings.Repeat(label63+".", 3) + strings.Repeat("a", 62), false}, // 254
	}
	for _, tt := range tests {
		err := ValidateSession(tt.name)
		if tt.ok && err != nil {
			t.Errorf("ValidateSession(%q) = %v, want nil", tt.name, err)
		}
		if !tt.ok && !errors.Is(err, ErrInvalidSession) {
			t.Errorf("ValidateSession(%q) = %v, want ErrInvalidSession", tt.name, err)
		}
	}
}

func TestValidateUploadPath(t *testing.T) {
	tests := []struct {
		path string
		ok   bool
	}{
		{"/handoff/Jane-resume.pdf", true},
		{"/handoff/sub/letter.pdf", true},
		{"/handoff/./a.pdf", true},
		{"/handoff/x/../a.pdf", true},
		{"/handoff/../etc/passwd", false},
		{"/handoff/a/../../etc/passwd", false},
		{"/handoff", false},
		{"/handoff/", false},
		{"/handoff/.", false},
		{"/handoffx/a.pdf", false},
		{"handoff/a.pdf", false},
		{"./a.pdf", false},
		{"", false},
		{"/etc/passwd", false},
		{"//handoff/../etc", false},
	}
	for _, tt := range tests {
		err := ValidateUploadPath(tt.path)
		if tt.ok && err != nil {
			t.Errorf("ValidateUploadPath(%q) = %v, want nil", tt.path, err)
		}
		if !tt.ok && !errors.Is(err, ErrUploadOutsideHandoff) {
			t.Errorf("ValidateUploadPath(%q) = %v, want ErrUploadOutsideHandoff", tt.path, err)
		}
	}
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{"no args", nil, nil},
		{"open", []string{"open", "https://careers.who.int"}, nil},
		{"upload ok", []string{"upload", "@e1", "/handoff/a.pdf", "/handoff/b.pdf"}, nil},
		{"flag before upload", []string{"--json", "upload", "@e1", "/handoff/a.pdf"}, nil},
		{"upload one bad", []string{"upload", "@e1", "/handoff/a.pdf", "/etc/passwd"},
			ErrUploadOutsideHandoff},
		{"upload traversal", []string{"upload", "@e1", "/handoff/../etc/passwd"},
			ErrUploadOutsideHandoff},
		{"upload prefix", []string{"upload", "@e1", "/handoffx/a.pdf"}, ErrUploadOutsideHandoff},
		{"upload dir", []string{"upload", "@e1", "/handoff"}, ErrUploadOutsideHandoff},
		{"upload no file", []string{"upload", "@e1"}, ErrUploadOutsideHandoff},
		{"upload bare", []string{"upload"}, ErrUploadOutsideHandoff},
		{"session flag", []string{"--session", "x", "open"}, ErrInvalidSession},
		{"session eq", []string{"--session=x", "open"}, ErrInvalidSession},
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
