// Package browse runs agent-browser (Vercel Labs) for `caddie browse`
// (PRD §7, §9; ADRs 0002, 0005). Browse validates the session name
// (ValidateSession) and the arguments (ValidateArgs: no --session override,
// upload files only under /handoff), builds the argv for the compose or
// local mode (Argv), applies CADDIE_BROWSE_TIMEOUT and runs it through the
// Exec interface; OSExec is the real os/exec implementation.
package browse
