package setup

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Ask puts each question to the user, one line per answer, and re-asks
// until the answer fits. At end of input every remaining question takes
// its default; a required question without one is an error.
func Ask(qs []Question, in io.Reader, out io.Writer) (Values, error) {
	r := bufio.NewReader(in)
	v := Values{}
	for _, q := range qs {
		val, err := askOne(q, r, out)
		if err != nil {
			return nil, err
		}
		v[q.ID] = val
	}
	return v, nil
}

func askOne(q Question, r *bufio.Reader, out io.Writer) (any, error) {
	for {
		_, _ = fmt.Fprint(out, prompt(q))
		line, err := r.ReadString('\n')
		eof := errors.Is(err, io.EOF)
		if err != nil && !eof {
			return nil, err
		}
		val, perr := Parse(q, line)
		if perr == nil {
			if eof {
				_, _ = fmt.Fprintln(out)
			}
			return val, nil
		}
		_, _ = fmt.Fprintf(out, "  %v\n", perr)
		if eof {
			return nil, perr
		}
	}
}

func prompt(q Question) string {
	var b strings.Builder
	b.WriteString(q.Prompt)
	if len(q.Options) > 0 {
		sep := " | "
		if q.Kind == KindChoices {
			sep = ", "
		}
		fmt.Fprintf(&b, " (%s)", strings.Join(q.Options, sep))
	}
	if q.Default != "" {
		fmt.Fprintf(&b, " [%s]", q.Default)
	}
	return b.String() + ": "
}
