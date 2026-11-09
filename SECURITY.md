# Security Policy

This policy covers the BrainBlitz backend, Telegram bot, shared libraries,
and deployment files maintained in this repository.

## Supported versions

Security fixes target the current `main` branch. Older tags, release snapshots,
and development branches do not have a separate security maintenance or
backport commitment. When reporting a problem, include the affected commit SHA
or release tag; reproducing it on current `main` is helpful but is not required
to submit a report.

## Reporting a vulnerability

Please report suspected vulnerabilities privately. Do not publish exploit
details in an issue, pull request, or public chat.

1. Open the repository's [Security advisories page](https://github.com/nima-abdpoor/brain-blitz/security/advisories).
2. If **Report a vulnerability** is available, use it to submit a private report.
3. If that option is unavailable, open an [issue](https://github.com/nima-abdpoor/brain-blitz/issues/new)
   asking only for a private security contact. Do not include the vulnerability
   details, affected data, credentials, or proof of concept in that issue.

Include the following in the private report when available:

- A description of the issue and its security impact.
- Affected services, endpoints, files, and commit SHA or version.
- Reproduction steps or a minimal proof of concept using test accounts.
- Relevant configuration and redacted logs or request/response examples.
- Any proposed mitigation or fix.

Reports involving authentication, authorization, access to other players' data,
credential exposure, game-state manipulation across accounts, or resource
exhaustion are welcome. Dependency vulnerabilities are also relevant when you
can describe how the affected dependency is used here.

Do not include live passwords, JWTs, Telegram bot tokens, private `.env` files,
or other users' personal data. Use a local or explicitly authorized test
deployment for reproduction, and avoid disrupting other players or systems.

## Handling and disclosure

Maintainers review reports on a best-effort basis. There is no guaranteed
response or release timeline. During review, maintainers may ask for additional
details, confirm the affected versions, and collaborate on a fix or workaround.

Please coordinate publication of technical details with maintainers so users
can receive a fix or mitigation. Confirmed issues may be documented in a
GitHub security advisory. Let maintainers know whether and how you would like
to be credited.

Ordinary bugs and feature requests should use the public issue templates and
the [contributor guidelines](CONTRIBUTING.md).

## Maintainer setup on GitHub

Adding this file does not enable GitHub's private vulnerability reporting.
A repository administrator can enable it under **Settings → Advanced Security
→ Private vulnerability reporting**, following
[GitHub's setup instructions](https://docs.github.com/en/code-security/how-tos/report-and-fix-vulnerabilities/configure-vulnerability-reporting/configure-for-a-repository).

For local secrets, network exposure, and deployment requirements, see
[the deployment guide](docs/run-on-another-computer.md).
