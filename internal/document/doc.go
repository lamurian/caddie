// Package document checks the agent's Markdown drafts before rendering:
// Verify traces every factual claim to history/ or answers.yml and records
// verified content hashes in .caddie/verified/; Keywords reports how well a
// draft covers a posting's requirements (PRD §6.2–6.3, ADR 0004).
package document
