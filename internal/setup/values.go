package setup

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ErrAnswer marks a setup answer that does not fit its question.
var ErrAnswer = errors.New("invalid setup answer")

// Values are parsed answers keyed by question ID: a string (text and
// choice), an int, or a []string (list and choices).
type Values map[string]any

// Parse reads what a user typed for q; empty text means q's default.
func Parse(q Question, raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = q.Default
	}
	if raw == "" && q.Required {
		return nil, fmt.Errorf("%w: %s: an answer is required", ErrAnswer, q.ID)
	}
	switch q.Kind {
	case KindInt:
		return parseInt(q, raw)
	case KindList, KindChoices:
		return checkList(q, splitList(raw))
	case KindChoice:
		if !slices.Contains(q.Options, raw) {
			return nil, fmt.Errorf("%w: %s: %q is not one of %s", ErrAnswer, q.ID, raw, strings.Join(q.Options, "|"))
		}
	}
	return raw, nil
}

func parseInt(q Question, raw string) (any, error) {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return nil, fmt.Errorf("%w: %s: %q is not a whole number of 0 or more", ErrAnswer, q.ID, raw)
	}
	return n, nil
}

func splitList(raw string) []string {
	out := []string{}
	for item := range strings.SplitSeq(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

// checkList checks a list answer: every item is one of a choices
// question's options, at most once, and a required list is not empty.
func checkList(q Question, items []string) (any, error) {
	if q.Required && len(items) == 0 {
		return nil, fmt.Errorf("%w: %s: an answer is required", ErrAnswer, q.ID)
	}
	if q.Kind != KindChoices {
		return items, nil
	}
	for i, it := range items {
		if !slices.Contains(q.Options, it) {
			return nil, fmt.Errorf("%w: %s: %q is not one of %s", ErrAnswer, q.ID, it, strings.Join(q.Options, ", "))
		}
		if slices.Contains(items[:i], it) {
			return nil, fmt.Errorf("%w: %s: %q is listed twice", ErrAnswer, q.ID, it)
		}
	}
	return items, nil
}

// FromJSON reads answers decoded from JSON (an agent's --answers file):
// strings are parsed like typed text, numbers fill int questions, and
// arrays of strings fill list questions. A missing ID takes its default;
// an unknown ID is an error, so a typo cannot pass silently.
func FromJSON(qs []Question, in map[string]any) (Values, error) {
	known := map[string]bool{}
	out := Values{}
	for _, q := range qs {
		known[q.ID] = true
		v, err := fromJSON(q, in[q.ID])
		if err != nil {
			return nil, err
		}
		out[q.ID] = v
	}
	for id := range in {
		if !known[id] {
			return nil, fmt.Errorf("%w: unknown question id %q", ErrAnswer, id)
		}
	}
	return out, nil
}

func fromJSON(q Question, v any) (any, error) {
	switch v := v.(type) {
	case nil:
		return Parse(q, "")
	case string:
		return Parse(q, v)
	case float64:
		if q.Kind == KindInt && v == float64(int(v)) {
			return parseInt(q, strconv.Itoa(int(v)))
		}
	case []any:
		if q.Kind == KindList || q.Kind == KindChoices {
			return listFromJSON(q, v)
		}
	}
	return nil, fmt.Errorf("%w: %s: %v does not fit a %s question", ErrAnswer, q.ID, v, q.Kind)
}

func listFromJSON(q Question, v []any) (any, error) {
	items := make([]string, 0, len(v))
	for _, it := range v {
		s, ok := it.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s: list items must be text, got %v", ErrAnswer, q.ID, it)
		}
		if s = strings.TrimSpace(s); s != "" {
			items = append(items, s)
		}
	}
	return checkList(q, items)
}
