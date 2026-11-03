package browse

import (
	"errors"
	"testing"
)

func TestValidateArgsRefusesBatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"batch upload escape", []string{"batch", "upload @e1 /etc/passwd"}},
		{"batch screenshot escape", []string{"batch", "screenshot /handoff/x.png"}},
		{"batch from stdin", []string{"batch"}},
		{"batch json stdin", []string{"--json", "batch"}},
		{"batch harmless", []string{"batch", "open https://careers.who.int", "snapshot -i"}},
		{"batch as value", []string{"fill", "@e1", "batch"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateArgs(tt.args); !errors.Is(err, ErrBatchRefused) {
				t.Fatalf("err = %v, want ErrBatchRefused", err)
			}
		})
	}
}

func TestValidateArgsRefusesFlags(t *testing.T) {
	for _, f := range refusedFlags {
		for _, args := range [][]string{
			{f, "x", "open", "https://careers.who.int"},
			{f + "=x", "snapshot"},
			{"snapshot", f},
		} {
			if err := ValidateArgs(args); !errors.Is(err, ErrFlagRefused) {
				t.Errorf("ValidateArgs(%q) = %v, want ErrFlagRefused", args, err)
			}
		}
	}
}

func TestValidateArgsAllowsLookalikes(t *testing.T) {
	for _, args := range [][]string{
		{"fill", "@e1", "batch processing"},
		{"fill", "@e1", "--profiles"},
		{"snapshot", "-i", "--json"},
		{"state", "save", "/trace/s.json"},
	} {
		if err := ValidateArgs(args); err != nil {
			t.Errorf("ValidateArgs(%q) = %v, want nil", args, err)
		}
	}
}
