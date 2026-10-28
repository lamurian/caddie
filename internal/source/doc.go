// Package source queries the job boards (Adzuna, Remotive, Arbeitnow,
// Greenhouse and Lever boards), normalises their results to Job and fetches
// single postings for `caddie posting save`. HTTP goes through Doer so
// tests replay recorded responses from testdata/sources/. It never touches
// the workspace (PRD §7, ADR 0016).
package source
