# Security policy

## Supported versions

| Version | Supported |
| ------- | --------- |
| latest `v0.x` tag | yes |
| older tags | no |

ownframe is pre-1.0. Only the latest tagged release gets security
fixes, which land as a new tag. Older tags stay available but unfixed;
the `retract` policy in `documentation/supply-chain.md` explains why a
bad tag is never re-tagged.

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability. Use GitHub
Private Vulnerability Reporting: the repository Security tab > Report a
vulnerability. That opens a private draft security advisory visible
only to the reporter and the maintainers.

What to include: the affected tag or commit, a minimal template or
program that triggers the issue, what you expected, and what happened
instead. A crash log or a failing test helps more than a description.

## Response SLA

- Acknowledgement within 3 business days.
- Triage (valid or not, severity, affected versions) within 7 days.
- Fix or mitigation plan within 30 days for high severity, 90 days
  otherwise. Timelines slip when a fix needs an engine-side change;
  the advisory stays updated in that case.

These are targets, not guarantees; the project has no paid support.

## Disclosure and the Go vulnerability database

Accepted reports become GitHub Security Advisories (GHSA). Publishing
the GHSA is what feeds the Go vulnerability database, so downstream
`govulncheck` runs flag the affected versions automatically. The
advisory stays private until a fixed tag exists, then goes public with
credit to the reporter unless they decline.

Researchers: web previews (`Serve`), devtools, debug endpoints, and
the desktop-cat notify port are covered. Social engineering, physical
access, and denial of service against the demo deployments are out of
scope.
