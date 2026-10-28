package workspace

import (
	"errors"
	"fmt"
)

// Status is a posting's lifecycle state (ADR 0011).
type Status string

// The posting states.
const (
	StatusSaved    Status = "saved"
	StatusTailored Status = "tailored"
	StatusRendered Status = "rendered"
	StatusApplied  Status = "applied"
	StatusManual   Status = "manual"
)

// Statuses lists every state in lifecycle order.
var Statuses = []Status{StatusSaved, StatusTailored, StatusRendered, StatusApplied, StatusManual}

// Sentinel errors for the state machine.
var (
	ErrUnknownStatus     = errors.New("unknown posting status")
	ErrIllegalTransition = errors.New("illegal status transition")
)

// transitions: saved → tailored → rendered → applied, and any pre-applied
// state → manual. applied and manual are terminal.
var transitions = map[Status][]Status{
	StatusSaved:    {StatusTailored, StatusManual},
	StatusTailored: {StatusRendered, StatusManual},
	StatusRendered: {StatusApplied, StatusManual},
}

// ParseStatus validates a status string.
func ParseStatus(s string) (Status, error) {
	for _, st := range Statuses {
		if string(st) == s {
			return st, nil
		}
	}
	return "", fmt.Errorf("%w: %q (want saved|tailored|rendered|applied|manual)", ErrUnknownStatus, s)
}

// Transition returns nil if from → to is legal, else an error wrapping
// ErrIllegalTransition or ErrUnknownStatus.
func Transition(from, to Status) error {
	for _, s := range []Status{from, to} {
		if _, err := ParseStatus(string(s)); err != nil {
			return err
		}
	}
	for _, next := range transitions[from] {
		if next == to {
			return nil
		}
	}
	return fmt.Errorf("%w: %s → %s", ErrIllegalTransition, from, to)
}
