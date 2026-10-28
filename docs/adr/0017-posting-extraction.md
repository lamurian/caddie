# 0017. Posting fetch and section extraction

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §6.1, §12.3, §12.4, D19

## Context

`caddie posting save <url>` must turn any job page into
`postings/<slug>.md` with the four sections of PRD §12.3, deterministically
and without an LLM. Pages differ widely: some boards have APIs, many
careers pages embed schema.org data, and the rest are plain HTML.

## Decision

- **Fetch order.** Greenhouse (`*.greenhouse.io/<board>/jobs/<id>`) and
  Lever (`jobs[.eu].lever.co/<board>/<uuid>`) URLs go through their public
  JSON APIs. Any other page is read for a schema.org `JobPosting` in
  JSON-LD (object, array or `@graph`), which gives title, company,
  location, salary, `validThrough` as deadline and the description. Without
  it, the title is `og:title`, `<h1>` or `<title>`, the company is
  `og:site_name` or the host, and the description is `<main>`, `<article>`
  or `<body>`. Scripts, styles, forms, buttons and `<nav>` are dropped.
  A page with no title is exit 5.
- **HTML parsing** uses `golang.org/x/net/html`, the Go team's HTML5
  parser; regular expressions cannot parse real pages reliably.
- **Headings** are `h1`–`h6`, a paragraph whose text is all bold, or a
  paragraph of at most 80 characters ending in a colon.
- **Sections:** each heading is matched, in this order, against cue
  phrases for Perks and benefits ("benefit", "we offer"…), Requirements
  ("requirement", "qualification", "bring", "nice to have"…) and
  Responsibilities ("responsibilit", "what you'll do", "duties"…). Text
  before the first recognised heading, and under any heading no cue
  matches, goes to Description. Original sub-headings stay as bold lines;
  a heading equal to the section name or the job title is dropped. An
  empty section reads `_Not stated in the posting._` and is listed in
  `missing_sections`.
- **Merging:** the last search's ID, source, title and company win, so the
  file matches what the user confirmed; other fields come from the page,
  then from the search.
- **Slugs** are `<title>-at-<company>` (ADR 0014). If that file exists in
  `postings/`, `applied/` or `manual/` for another URL, `-2`, `-3`… is
  appended. Saving a URL that is already saved rewrites nothing and returns
  its slug, so the command is safe to repeat.
- **Reviews:** `reviews/<company-slug>.md`, else the review whose
  frontmatter `company` matches, has its body copied under `## Reviews`;
  otherwise the section names the file to create. `ats` is set for known
  applicant tracking hosts and `ats_host` is always the URL's host.

## Consequences

- Most postings land with usable sections, but a page whose headings use
  unexpected words puts everything under Description; the agent still
  reads the whole file, and new cues are a one-line change with a test.
- Trailing text such as "Application deadline: …" stays in the last
  section it follows.
- Page text is stored as data and never interpreted (AGENTS.md rule 5).
