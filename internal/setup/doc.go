// Package setup creates preferences.yml and answers.yml on first run: the
// default question set (shared by `caddie init` and the agent), parsing
// and checking answers, rendering both files and validating them with the
// loaders every command uses. See docs/adr/0032-setup-init-and-schemas.md.
package setup
