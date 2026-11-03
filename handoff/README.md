# handoff/

The **only** directory the sandboxed browser can see. The browser container
mounts it read-only at `/handoff`, so a file must be here to be uploaded to
an application form ([ADR 0002](../docs/adr/0002-sandboxed-browser.md)).

## Lifecycle

1. `caddie render resume|cover-letter <slug>` writes the PDF to `resumes/` or
   `cover-letters/` and copies it here.
2. The agent uploads it: `caddie browse <host> upload <selector> /handoff/<file>.pdf`.
   Paths outside `/handoff`, relative paths and `..` are refused (exit 7).
3. `caddie move <slug> applied|manual` clears the directory.

It holds only the PDFs for the application in progress. Keeping it separate
from `resumes/` means the browser never sees drafts, history or answers.

## Contributing

- Runtime state: everything except this README and `.gitignore` is ignored
  and blocked by `scripts/guard-workspace.sh`
  ([ADR 0024](../docs/adr/0024-workspace-readmes.md)).
- Never copy files here by hand to get around the verified-render gate.
