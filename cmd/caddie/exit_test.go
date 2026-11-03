package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func exitTable(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok1 := strings.Cut(string(data), "<!-- exit-codes:begin -->")
	table, _, ok2 := strings.Cut(rest, "<!-- exit-codes:end -->")
	if !ok1 || !ok2 {
		t.Fatalf("%s: exit-code markers not found", path)
	}
	return strings.TrimSpace(table)
}

// The skill and its exit-code reference must carry the exit-code table
// verbatim (cmd/README.md).
func TestExitTableMirroredInSkill(t *testing.T) {
	readme := exitTable(t, "../README.md")
	for _, path := range []string{"../../skills/caddie/SKILL.md", "../../skills/caddie/references/exit-codes.md"} {
		if got := exitTable(t, path); got != readme {
			t.Errorf("exit-code tables differ:\ncmd/README.md:\n%s\n\n%s:\n%s", readme, path, got)
		}
	}
	for _, code := range []int{exitOK, exitCheckFailed, exitUsage, exitConfig,
		exitMalformed, exitNotFound, exitIllegalState, exitRefused, exitExternal, exitInternal} {
		if !strings.Contains(readme, fmt.Sprintf("\n| %d |", code)) {
			t.Errorf("exit code %d is not documented in cmd/README.md", code)
		}
	}
}
