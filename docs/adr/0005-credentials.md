# 0005. Credentials never enter the agent's context

- **Status:** Accepted
- **Date:** 2026-10-28
- **PRD refs:** §9, §10, D10

## Context

Some portals require accounts. Anything in the agent's context can leak
through logs, model providers or prompt injection.

## Decision

- Public forms need no login.
- Password portals: the user stores the credential once with agent-browser's
  credential store (`auth save … --password-stdin`). Afterwards the agent
  only calls `auth login <name>`.
- 2FA and OAuth: the user relays a one-time code through chat. As a fallback,
  use a headed `state save` on a laptop followed by `state load`.
- Sessions are named by hostname, and `caddie browse` validates the name.
- Session state files hold plaintext tokens, so they live only in the
  browser volume (0700) and are never bind-mounted into the repo.
- The skill forbids entering or requesting a password in chat.

## Consequences

- Initial setup per portal needs a one-time manual step by the user.
- `*.state.json` and `docker/browser-profile/` are git-ignored as a backstop.
- No secrets in `SKILL.md`, postings, resumes, letters, answers or traces.
