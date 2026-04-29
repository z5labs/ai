# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This repo contains Claude Code agents and skills maintained by z5labs. It is not a traditional application codebase — it holds markdown-defined agent definitions and skill templates that extend Claude Code's capabilities.

## Repository Structure

- `agents/` — Agent definitions (markdown files with frontmatter). Each agent describes a specialized subprocess with specific tools, model, and instructions.
- `plugins/` — Plugin packages, each containing a `.claude-plugin/plugin.json` manifest and a `skills/` subdirectory. Each skill lives in its own subdirectory containing a `SKILL.md` file that defines the skill's behavior and scaffolding instructions.

## Key Concepts

### File Library Pattern
The primary domain is **file library packages** — Go packages for parsing and formatting file formats. Two pipeline shapes are supported:

- **Text formats** follow **Tokenizer → Parser → AST → Printer**.
  - Scaffold: `/new-go-text-file-library [name]` (lives in the `file-library` plugin)
  - Implementer agent: `implement-go-text-file-library` (test-first: tokenizer tests → tokenizer → parser tests → parser → printer tests → printer)
- **Binary formats** follow **Types → Decoder → Encoder**.
  - Scaffold: `/new-go-binary-file-library [name]` (lives in the `file-library` plugin)
  - Implementer agent: `implement-go-binary-file-library`

Skill and agent names carry the `-go-` language prefix so the orchestration shape can be reused across other languages later.

### Go Conventions (for file library packages)
- State machine pattern with recursive action functions for tokenizer, parser, and printer
- `iter.Seq2` / `iter.Pull2` for token streaming
- Table-driven tests with `t.Parallel()` and `testify/require`
- Complex types must use the inner action loop pattern (no inline for-loops)
- Run `go test -race ./...` to verify changes

## Commit Convention

Commits use the format: `scope: description` (e.g., `ai: implement file library agent`). Use lowercase descriptions.

## License

MIT (z5labs)
