// Package posting saves a job posting to postings/<slug>.md: it fetches the
// page through internal/source, sorts the description into Description,
// Responsibilities, Requirements and Perks and benefits, writes the PRD
// §12.3 frontmatter with status saved, copies reviews/<company>.md into
// ## Reviews and marks the job saved in seen.jsonl (ADR 0016, ADR 0017).
package posting
