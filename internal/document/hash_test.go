package document

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// A recorded hash vouches for exactly that content: a one-byte edit
// needs a new verify.
func TestVerifiedHash(t *testing.T) {
	state := filepath.Join(t.TempDir(), ".caddie")
	content := readDraft(t, "resume-traceable.md")
	if ok, err := IsVerified(state, content); ok || err != nil {
		t.Fatalf("before recording: %v %v", ok, err)
	}
	hash, err := RecordVerified(state, "resumes/x.md", content, "2026-10-29")
	if err != nil || hash != Hash(content) || len(hash) != 64 {
		t.Fatalf("RecordVerified = %q, %v", hash, err)
	}
	if ok, err := IsVerified(state, content); !ok || err != nil {
		t.Errorf("after recording: %v %v", ok, err)
	}
	edited := append([]byte{}, content...)
	edited[len(edited)-2] ^= 1
	if ok, err := IsVerified(state, edited); ok || err != nil {
		t.Errorf("one-byte edit still verified: %v %v", ok, err)
	}
}

// The verified/<hash> file names the draft and the date; recording the
// same content again replaces it.
func TestVerifiedRecord(t *testing.T) {
	state := t.TempDir()
	content := []byte("# Draft\n")
	for _, at := range []string{"2026-10-29", "2026-10-30"} {
		if _, err := RecordVerified(state, "resumes/x.md", content, at); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(state, VerifiedDir, Hash(content)))
	if err != nil {
		t.Fatal(err)
	}
	var rec verifiedRecord
	if err := json.Unmarshal(data, &rec); err != nil || rec.Path != "resumes/x.md" || rec.At != "2026-10-30" {
		t.Errorf("record %s: %+v %v", data, rec, err)
	}
	if entries, _ := os.ReadDir(filepath.Join(state, VerifiedDir)); len(entries) != 1 {
		t.Errorf("verified/ holds %d files, want 1", len(entries))
	}
}

// A record names the draft relative to the workspace root, slash-separated,
// or by its absolute path when it lies outside; relative arguments resolve
// against the current directory.
func TestRecordPath(t *testing.T) {
	t.Chdir(t.TempDir())
	base, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	ws := filepath.Join(base, "ws")
	cases := []struct{ root, doc, want string }{
		{".", "./resumes/x.md", "resumes/x.md"},
		{".", "resumes/x.md", "resumes/x.md"},
		{".", filepath.Join(base, "resumes", "x.md"), "resumes/x.md"},
		{base, "cover-letters/x.md", "cover-letters/x.md"},
		{".", "..x.md", "..x.md"},
		{"ws", "ws/resumes/x.md", "resumes/x.md"},
		{"./ws/", "ws/./resumes/../resumes/x.md", "resumes/x.md"},
		{ws, filepath.Join(ws, "resumes", "x.md"), "resumes/x.md"},
		{"ws", "resumes/x.md", filepath.Join(base, "resumes", "x.md")},
		{"ws", "ws2/x.md", filepath.Join(base, "ws2", "x.md")},
		{ws, "../outside/x.md", filepath.Join(filepath.Dir(base), "outside", "x.md")},
	}
	for _, c := range cases {
		if got, err := RecordPath(c.root, c.doc); got != c.want || err != nil {
			t.Errorf("RecordPath(%q, %q) = %q, %v; want %q", c.root, c.doc, got, err, c.want)
		}
	}
}
