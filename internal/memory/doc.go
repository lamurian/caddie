// Package memory is Caddie's plain-file memory: in the state directory,
// seen.jsonl (which jobs were shown, rejected or saved) and the
// last-search.json cache that lets `seen add` and `posting save` recover a
// job from its ID or URL; and answers.yml, the user's stored answers in
// three scopes (PRD §12.5, ADR 0007, ADR 0016, ADR 0020). Files are read
// leniently and every unusable line is reported.
package memory
