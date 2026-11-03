# Isolated and trusted browser modes

## Decision

Snaggle restores upstream content features while separating browser authority:

- normal mode supports public URL batches, Markdown, HTML, text, PDF, PNG,
  metadata, file output, custom user agents, and positive timeouts;
- normal mode always launches sandboxed headless Chromium with an ephemeral
  profile, a random loopback CDP port, and the public-network egress proxy;
- `--trusted-session` is required for visible browsers, persistent profiles,
  fixed CDP ports, and existing-tab access;
- process killing, local files, private-network access, update checks,
  telemetry, and runtime skill catalogs remain disabled.

## Rationale

Prompt injection is shared with ordinary web tools and is handled by the agent
trust boundary. Local-network reachability and browser-session reuse are unique
local-browser capabilities, so they remain isolated or require explicit intent.
Output formats and Markdown conversion do not expand those capabilities.
