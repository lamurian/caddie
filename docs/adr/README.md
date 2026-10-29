# Architecture Decision Records

Decisions that shape Caddie, derived from the PRD's decision table
([PROPOSAL.md §4](../../PROPOSAL.md#4-decisions-and-rationale)). Agents must
read the ADRs a task lists before starting it, and must not contradict an
accepted ADR without superseding it.

| ADR | Title | PRD |
|-----|-------|-----|
| [0001](0001-skill-plus-cli.md) | Skill + single Go CLI; judgment vs mechanics | D1–D3 |
| [0002](0002-sandboxed-browser.md) | Sandboxed headless browser behind an egress allowlist | D4–D7 |
| [0003](0003-human-submit-gate.md) | Human submit gate | D8 |
| [0004](0004-fidelity-check.md) | Mandatory fidelity check before rendering | D9 |
| [0005](0005-credentials.md) | Credentials never enter the agent's context | D10 |
| [0006](0006-portals-and-manual.md) | Portal adapters and the manual checklist | D11–D12 |
| [0007](0007-plain-file-memory.md) | Plain-file, user-editable memory | D13 |
| [0008](0008-workspace-tracking.md) | Tracked workspace dirs; public history, private contact | D14–D15 |
| [0009](0009-history-schema.md) | History schema and identity source | D16–D17 |
| [0010](0010-quarto-rendering.md) | PDF rendering with Quarto + Typst | D18 |
| [0011](0011-posting-state-machine.md) | Posting frontmatter as a state machine; MIT licence | D19–D20 |
| [0012](0012-agentic-docs-and-line-limits.md) | Agentic docs and line limits | — |
| [0013](0013-dev-tooling.md) | Dev tooling: mise, lefthook and Renovate | — |
| [0014](0014-cli-contract.md) | CLI contract: exit codes, output, configuration | §7, §12.7–12.8 |
| [0015](0015-history-parser-details.md) | History parser details | §12.1–12.2 |
| [0016](0016-job-ids-and-search-memory.md) | Job IDs, search filtering and search memory | §6.1, §7, §12.5 |
| [0017](0017-posting-extraction.md) | Posting fetch and section extraction | §6.1, §12.3–12.4 |
| [0018](0018-board-matching-and-fair-limit.md) | Company boards match titles; `--limit` shared across feeds | §6.1, §7 |
| [0019](0019-portal-config-schema.md) | Portal config schema, IDs and enabling | §7, §11, D11 |
| [0021](0021-claim-extraction-and-non-factual-prose.md) | Claim extraction and non-factual prose | §6.2–6.3, §13, D9 |

## Writing a new ADR

1. Copy [`template.md`](template.md) to `NNNN-short-title.md` using the next
   number.
2. Fill it in: under 100 lines, one decision, consequences stated honestly.
3. Add it to the table above.
4. To reverse a decision, write a new ADR. Then set the old one's status to
   `Superseded by NNNN`; never delete or rewrite it.

Statuses: `Proposed` → `Accepted` → (`Deprecated` | `Superseded by NNNN`).
