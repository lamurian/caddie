package workspace

import (
	"errors"
	"testing"
)

// legal is the complete set of legal transitions from ADR 0011.
var legal = map[[2]Status]bool{
	{StatusSaved, StatusTailored}:    true,
	{StatusTailored, StatusRendered}: true,
	{StatusRendered, StatusApplied}:  true,
	{StatusSaved, StatusManual}:      true,
	{StatusTailored, StatusManual}:   true,
	{StatusRendered, StatusManual}:   true,
}

func TestEveryLegalTransition(t *testing.T) {
	for pair := range legal {
		if err := Transition(pair[0], pair[1]); err != nil {
			t.Errorf("%s → %s: unexpected error %v", pair[0], pair[1], err)
		}
	}
}

func TestAllOtherTransitionsAreIllegal(t *testing.T) {
	for _, from := range Statuses {
		for _, to := range Statuses {
			if legal[[2]Status{from, to}] {
				continue
			}
			if err := Transition(from, to); !errors.Is(err, ErrIllegalTransition) {
				t.Errorf("%s → %s: got %v, want ErrIllegalTransition", from, to, err)
			}
		}
	}
}

func TestIllegalSkipToApplied(t *testing.T) {
	err := Transition(StatusSaved, StatusApplied)
	if !errors.Is(err, ErrIllegalTransition) || err.Error() != "illegal status transition: saved → applied" {
		t.Errorf("got %v", err)
	}
}

func TestUnknownStatus(t *testing.T) {
	if err := Transition("drafted", StatusTailored); !errors.Is(err, ErrUnknownStatus) {
		t.Errorf("from: got %v", err)
	}
	if err := Transition(StatusSaved, "submitted"); !errors.Is(err, ErrUnknownStatus) {
		t.Errorf("to: got %v", err)
	}
	if _, err := ParseStatus("Saved"); !errors.Is(err, ErrUnknownStatus) {
		t.Errorf("status is case-sensitive: got %v", err)
	}
}
