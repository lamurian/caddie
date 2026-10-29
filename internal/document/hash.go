package document

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// VerifiedDir is the directory in the state directory (`.caddie/`) that
// holds one file per verified content hash (ADR 0004).
const VerifiedDir = "verified"

// verifiedRecord is the body of a verified/<hash> file: which draft
// passed (as made by RecordPath), and when. Only the file name (the hash)
// is authoritative; Path is informational.
type verifiedRecord struct {
	Path string `json:"path"`
	At   string `json:"at"`
}

// Hash is the hex SHA-256 of content.
func Hash(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// RecordPath is the path a verified record names for the draft at
// docPath: relative to the workspace root and slash-separated
// ("resumes/<slug>.md"), or the cleaned absolute path when the draft lies
// outside the workspace. Relative arguments resolve against the current
// directory.
func RecordPath(root, docPath string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absDoc, err := filepath.Abs(docPath)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, absDoc)
	if err != nil {
		return absDoc, nil //nolint:nilerr // e.g. another volume: outside the workspace
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return absDoc, nil
	}
	return rel, nil
}

// RecordVerified marks content as verified by writing
// <stateDir>/verified/<hash>, naming the draft's path (as made by
// RecordPath) and the date (at). Call it only after Verify reports OK.
func RecordVerified(stateDir, docPath string, content []byte, at string) (string, error) {
	hash := Hash(content)
	dir := filepath.Join(stateDir, VerifiedDir)
	data, err := json.Marshal(verifiedRecord{Path: docPath, At: at})
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	tmp := filepath.Join(dir, hash+".tmp")
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return "", err
	}
	return hash, os.Rename(tmp, filepath.Join(dir, hash))
}

// IsVerified reports whether this exact content was verified. Any edit,
// even one byte, changes the hash, so the draft must be verified again.
func IsVerified(stateDir string, content []byte) (bool, error) {
	_, err := os.Stat(filepath.Join(stateDir, VerifiedDir, Hash(content)))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}
