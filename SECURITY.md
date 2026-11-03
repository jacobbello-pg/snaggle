# Security Policy

This private repository is maintained for personal use. Report suspected issues
privately to the repository owner; do not include credentials, cookies, tokens,
private page content, or browser profiles in an issue.

## Supported branch

Only the current `main` branch is supported.

## Security model

Snaggle renders untrusted public pages in an ephemeral, sandboxed Chromium
process and returns bounded text to an AI workflow. Page output remains
untrusted and can contain prompt injection.

The following are outside the supported boundary: authenticated pages, local or
private-network targets, persistent profiles, existing browser sessions,
arbitrary file output, and browser use outside an isolated Chromium process.
