# ai

Claude Code agents and skills maintained by [z5labs](https://github.com/z5labs).

## Agents

| Agent | Description |
|-------|-------------|
| `implement-go-text-file-library` | Implements features for Go packages that follow a tokenizer/parser/printer pipeline. |
| `implement-go-binary-file-library` | Implements features for Go packages that follow a types/decoder/encoder pipeline. |

## Skills

| Skill | Description |
|-------|-------------|
| `new-go-text-file-library` | Scaffolds a new Go text file library package with tokenizer, parser, printer, and tests. |
| `new-go-binary-file-library` | Scaffolds a new Go binary file library package with types, decoder, encoder, and tests. (Bundled in the `file-library` plugin.) |
| `audit-skill` | Statically audits a Claude Code skill against four quality objectives (idempotency, reproducibility, context management, strict definitions) and posts findings to a PR review or a report file. (Bundled in the `audit-skill` plugin.) |

## License

[MIT](LICENSE)
