# manual/

Postings the user finishes **by hand**. Caddie routes a posting here instead
of submitting it when it meets a login wall, a CAPTCHA, an unsupported portal
(most international organisations), an unverifiable confirmation, or when the
user replies `manual` at the submit gate
([ADR 0006](../docs/adr/0006-portals-and-manual.md)). This is a designed
outcome, not a failure.

## What a file holds

The posting `<slug>.md` (with `status: manual` and `manual_reason`), moved
together with its resume, cover letter and trace. The posting gains a
`## Manual steps` section of `- [ ]` items: what the agent learned about the
form, as a checklist.

## Lifecycle

1. The agent writes the checklist, then runs
   `caddie move <slug> manual --reason "<one line>"`.
2. The user applies on the portal and ticks the items.
3. `caddie follow-ups` lists entries with unticked items or a near deadline.

## Contributing

Runtime state: everything except this README and `.gitignore` is ignored
([ADR 0024](../docs/adr/0024-workspace-readmes.md)). Fixtures belong in
`testdata/`.
