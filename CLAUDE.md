# uictl — UniFi Network CLI

This project includes `uictl`, a CLI for managing UniFi controllers (Network + Protect APIs).

Run `uictl skills` for the full agent reference, or `uictl schema` to discover commands.

## Quick Reference

```bash
uictl <resource> <action> [flags]
```

**Always**: use `--fields` on reads, `--dry-run` before writes, `--json-input` for complex payloads.

**Never**: parse table output, omit `--yes` on destructive commands in non-interactive mode.

See [AGENTS.md](./AGENTS.md) for the complete specification.
