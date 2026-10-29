// Package document checks the agent's Markdown drafts before rendering:
// Verify traces every factual claim to history/ or answers.yml (ADR 0021)
// and RecordVerified stores verified content hashes in .caddie/verified/;
// Keywords reports how well a draft covers a posting's requirements
// (PRD §6.2–6.3, ADR 0004). Render turns a verified draft into a PDF with
// quarto behind the Quarto interface, names it with PDFName and copies it
// to handoff/ (PRD §12.6, ADR 0022).
package document
