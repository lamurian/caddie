package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// Every `caddie <command>` a workflow reference shows must be a real
// command. (SKILL.md's verbatim hard rules name later-phase commands.)
func TestSkillCommandsExist(t *testing.T) {
	files, _ := filepath.Glob("../../skills/caddie/references/*.md")
	if len(files) == 0 {
		t.Fatal("no workflow references found")
	}
	cmdRef := regexp.MustCompile("`caddie ([a-z-]+)")
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range cmdRef.FindAllStringSubmatch(string(data), -1) {
			if _, ok := commands[m[1]]; !ok {
				t.Errorf("%s: `caddie %s` is not a command", f, m[1])
			}
		}
	}
}
