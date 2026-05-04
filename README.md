# ai

Claude Code agents and skills maintained by [z5labs](https://github.com/z5labs).

## Agents

| Agent | Description |
|-------|-------------|
| `file-library` | Autonomous orchestrator that runs the full file-library workflow end-to-end (extract → scaffold → implement → verify) against `go test -race`. (Bundled in the `file-library` plugin; works under both Claude Code and GitHub Copilot CLI.) |

## Skills

| Skill | Description |
|-------|-------------|
| `extract-text-spec` | Extracts a text file format specification (RFC, HTML, PDF) into a single progressive-disclosure `SPEC.md`. (Bundled in the `file-library` plugin.) |
| `extract-binary-spec` | Extracts a binary file format specification into a chunked, per-structure markdown reference. (Bundled in the `file-library` plugin.) |
| `new-go-text-file-library` | Scaffolds a new Go text file library package with tokenizer, parser, printer, and tests. (Bundled in the `file-library` plugin.) |
| `new-go-binary-file-library` | Scaffolds a new Go binary file library package with types, decoder, encoder, and tests. (Bundled in the `file-library` plugin.) |
| `implement-go-text-file-library` | Implements features for Go packages that follow a tokenizer/parser/printer pipeline. (Bundled in the `file-library` plugin.) |
| `implement-go-binary-file-library` | Implements features for Go packages that follow a types/decoder/encoder pipeline. (Bundled in the `file-library` plugin.) |
| `review-file-library` | Audits an existing Go file library package against its `SPEC.md`, reporting missing coverage and drift. (Bundled in the `file-library` plugin.) |
| `add-fixture` | Adds a real-world file as a golden test fixture and wires up a round-trip test. (Bundled in the `file-library` plugin.) |
| `audit-skill` | Statically audits a Claude Code skill against four quality objectives (idempotency, reproducibility, context management, strict definitions) and posts findings to a PR review or a report file. (Bundled in the `audit-skill` plugin.) |

## License

[MIT](LICENSE)
