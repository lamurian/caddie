# 0021. Claim extraction and non-factual prose

- **Status:** Accepted
- **Date:** 2026-10-29
- **PRD refs:** §6.2, §6.3, §13, D9

## Context

ADR 0004 makes `caddie verify` mandatory but leaves open what counts as a
claim and how prose such as "passionate about…" is treated. The check must
be deterministic, explain each failure by line, and err towards reporting.
It supersedes nothing; it closes ADR 0004's open point.

## Decision

`document.Verify` traces every claim of a Markdown draft to a corpus: the
profile (identity, about, section headings, entries, works, bullets), the
answer values in scope for the posting, and the targets (the posting's
company and title, which may be named but vouch for terms only). Matching
folds case and diacritics and compares whole words. Claim kinds:

- **entry**: each `###` block, the title plus up to three field lines
  directly after it (parts may be joined with `|`). The title must equal a
  history entry or work title; each field must equal that same entry's
  organisation, location or date range (a work's venue, authors or date).
  This catches upgraded titles and stretched or swapped dates even when
  every word or date exists elsewhere.
- **date**: `Mon YYYY`, full month names, `Sept`, `1 January 2027`,
  ISO dates and bare 19xx/20xx years, compared normalised ("Jun 2023" ==
  "June 2023"). A month-year must be in the history or answers (an
  answer's `YYYY-MM-DD` vouches for its month and year); a month where the
  history has only a year is untraceable (ADR 0015).
- **number**: digits outside dates, normalised ("40,000" == "40000",
  "30%" → 30, "07" → 7). The words two–twelve count as digits on both
  sides; "one" is ignored as it is mostly a pronoun.
- **term**: ALL-CAPS words, words with an inner capital, a digit or a
  symbol (PyTorch, C2, C++, Node.js), capitalised words that do not start a
  sentence, URLs and e-mail addresses. Adjacent term words form one phrase
  that must appear in that order ("Fjordline Clinical Institute" fails
  though each word exists). A sentence starts after a list marker, a blank
  line, a heading, `.`, `!` or `?` (not after `Dr.`, `e.g.`; not after
  `:`). A small allowlist covers letter conventions and calendar words
  (I, Dear, Regards, Hiring, Manager, Committee, months, weekdays, Present).
- **skill**: each item under a `#`/`##` heading containing *skills*,
  *tools*, *technologies*, *competencies* or *expertise*, split on `,`,
  `;`, bullets and " and " after an optional `Label:`, must appear as a
  phrase in the profile or answers. Targets do not vouch for skills.

`#` and `##` heading text is structure, not a claim; deeper headings are
prose. Frontmatter values are checked as prose with every capital counted;
keys are not claims.

**Non-factual prose:** a sentence with no claim token is non-factual and
passes ("passionate about open science"). Any claim token inside such a
sentence must still trace ("passionate about Kubernetes" fails). Opinion
and motivation need no answer; names, numbers, dates and skills do.

`RecordVerified` writes `.caddie/verified/<sha256>` (JSON: path, date)
only after a pass; `IsVerified` checks the exact bytes, so any edit needs a
new verify.

## Consequences

- Each failure names a line, a kind and the text, so the agent can fix or
  justify it; a stored answer is the way to add a fact not in `history/`.
- Known blind spots, which the tailor and cover-letter skill must cover by
  using the history's wording and never adding facts:
  - a lowercase skill or claim in prose outside the skills list ("built
    dashboards in tableau") and a capitalised invented word that starts a
    sentence are not claims;
  - recombining known words in lowercase or across sentences, and
    qualitative inflation ("led" for "supported", "expert in");
  - `##` heading text, and terms split across a line break, are checked
    word by word only;
  - words that happen to exist anywhere in the corpus pass in any context.
- False positives are accepted: a capital after `:`, contact details
  (phone, e-mail; they are merged at render time, ADR 0008) and a letter's
  own date must not be in the draft. A paraphrased skill fails until it
  uses the history's wording.
