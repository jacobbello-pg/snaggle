# Security Policy

This private repository is maintained for personal use. Report suspected issues
privately to the repository owner; do not include credentials, cookies, tokens,
private page content, or browser profiles in an issue.

## Supported branch

Only the current `main` branch is supported.

## Security model

Normal mode renders untrusted public pages in an ephemeral, sandboxed Chromium
process behind a public-network-only proxy. Markdown remains the default, and
stdout is bounded. Page output remains untrusted and can contain prompt
injection.

Profiles, visible browsers, fixed CDP ports, and existing browser sessions are
available only through the explicit `--trusted-session` mode. The embedded
agent skill requires direct user authorization before using that mode.

Local files, private-network targets, browser process killing, telemetry,
automatic update checks, and runtime skill catalogs remain outside the
supported boundary.
