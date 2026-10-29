package document

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func compactJSON(t *testing.T, s string) string {
	t.Helper()
	var b bytes.Buffer
	if err := json.Compact(&b, []byte(s)); err != nil {
		t.Fatalf("invalid JSON %q: %v", s, err)
	}
	return b.String()
}

func TestRenderQuartoFailure(t *testing.T) {
	var out strings.Builder
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&out, "progress line %02d\n", i)
	}
	out.WriteString("ERROR: typst compile failed\n")
	cases := map[string]*fakeQuarto{
		"exit error": {out: []byte(out.String()), err: errors.New("exit status 1")},
		"no PDF":     {out: []byte(out.String()), noPDF: true},
	}
	for name, q := range cases {
		t.Run(name, func(t *testing.T) {
			in, _, tmp := newRenderFixture(t, true)
			_, err := Render(context.Background(), q, in)
			if !errors.Is(err, ErrQuarto) {
				t.Fatalf("err = %v, want ErrQuarto", err)
			}
			msg := err.Error()
			if !strings.Contains(msg, "ERROR: typst compile failed") || strings.Contains(msg, "progress line 01") {
				t.Errorf("message should hold only the output tail:\n%s", msg)
			}
			assertNoOutput(t, in, tmp)
		})
	}
}

func TestRenderMissingTemplate(t *testing.T) {
	in, _, tmp := newRenderFixture(t, true)
	in.Template = filepath.Join(filepath.Dir(in.Template), "nope.qmd")
	q := &fakeQuarto{}
	if _, err := Render(context.Background(), q, in); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v, want not exist", err)
	}
	if q.calls != 0 {
		t.Errorf("quarto called %d times", q.calls)
	}
	assertNoOutput(t, in, tmp)
}

// TestRenderContactStaysInTempDir: after a render, the e-mail and phone are
// in the PDFs only; no other workspace file holds them and the temp dir
// that did is gone (ADR 0008).
func TestRenderContactStaysInTempDir(t *testing.T) {
	in, ws, tmp := newRenderFixture(t, true)
	q := &fakeQuarto{}
	if _, err := Render(context.Background(), q, in); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(q.files[DataFile], testEmail) || !strings.Contains(q.files[DataFile], testPhone) {
		t.Fatalf("contact missing from the render dir: %q", q.files[DataFile])
	}
	pdfs := 0
	for rel, content := range snapshot(ws) {
		if strings.HasSuffix(rel, ".pdf") {
			pdfs++
			continue
		}
		if strings.Contains(content, testEmail) || strings.Contains(content, testPhone) {
			t.Errorf("contact leaked into %s", rel)
		}
	}
	if pdfs != 2 {
		t.Errorf("found %d PDFs, want 2 (source dir and handoff)", pdfs)
	}
	if entries, _ := os.ReadDir(tmp); len(entries) > 0 {
		t.Errorf("temp render dir left behind: %v", entries)
	}
}

func TestStripFrontmatter(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"none", "# Title\n\nBody.\n", "# Title\n\nBody.\n"},
		{"dashes", "---\na: b\n---\n# Title\n", "# Title\n"},
		{"dots", "---\na: b\n...\n\nBody\n", "\nBody\n"},
		{"crlf", "---\r\na: b\r\n---\r\nBody\r\n", "Body\r\n"},
		{"unclosed", "---\na: b\nBody\n", "---\na: b\nBody\n"},
		{"rule later", "# T\n\n---\n\nBody\n", "# T\n\n---\n\nBody\n"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		if got := string(stripFrontmatter([]byte(c.in))); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestRenderDataOmitsEmpty(t *testing.T) {
	data, err := json.Marshal(newRenderData(RenderInput{Kind: KindCoverLetter}))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"kind":"cover-letter"}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestExecQuartoRunsInDir(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh")
	}
	dir := t.TempDir()
	out, err := ExecQuarto{Bin: sh}.Run(context.Background(), dir, "-c", "pwd; echo oops >&2")
	if err != nil {
		t.Fatal(err)
	}
	real, _ := filepath.EvalSymlinks(dir)
	if got := string(out); !strings.Contains(got, real) || !strings.Contains(got, "oops") {
		t.Errorf("output %q: want dir %s and stderr", got, real)
	}
	if _, err := (ExecQuarto{Bin: filepath.Join(dir, "missing")}).Run(context.Background(), dir); err == nil {
		t.Error("missing binary: want an error")
	}
}
