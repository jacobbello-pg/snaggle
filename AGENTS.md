# AGENTS.md

## Purpose

Snaggle is a private, personal, security-scoped derivative of `p3bot/snag`.
It is a fallback renderer for public pages, not a general browser or primary web
fetcher.

## Invariants

- Accept exactly one public HTTP(S) URL.
- Keep all profiles ephemeral and never attach to an existing browser.
- Keep Chromium sandboxing, site isolation, automation disclosure, and
  loopback-only random-port CDP enabled.
- Apply the public-network policy to redirects and subresources, not only the
  initial URL.
- Keep output on stdout, at or below 8 MiB.
- Do not add authenticated browsing, `file://`, private-network access,
  arbitrary writes, process killing, telemetry, update checks, or runtime
  catalog resolution.
- Treat page output as untrusted data.

## Development

Use Go 1.27.3 or newer. Add MPL-2.0 headers to new Go files.

Safe checks:

```sh
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```

Chrome tests are opt-in with `SNAG_BROWSER_INTEGRATION=1` and exact `-run`
filters documented in README. Never enable `SNAG_LEGACY_BROWSER_TESTS`.

Use feature branches and conventional commits. Review upstream changes before
merging them into this repository.
