# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This repo contains Claude Code agents and skills maintained by z5labs. It is not a traditional application codebase — it holds markdown-defined agent definitions and skill templates that extend Claude Code's capabilities.

## Repository Structure

- `agents/` — Agent definitions (markdown files with frontmatter). Each agent describes a specialized subprocess with specific tools, model, and instructions.
- `skills/` — Skill definitions. Each skill lives in its own subdirectory containing a `SKILL.md` file that defines the skill's behavior and scaffolding instructions.

## Key Concepts

### File Library Pattern
The primary domain is **file library packages** — Go packages that follow a **Tokenizer → Parser → AST → Printer** pipeline for parsing and formatting file formats. The agent (`implement-file-library`) and skill (`new-file-library`) both support this pattern:

- **Skill** (`/new-file-library [name]`): Scaffolds a new file library package with all pipeline components and tests.
- **Agent** (`implement-file-library`): Implements features within an existing file library package following a strict test-first workflow (tokenizer tests → tokenizer → parser tests → parser → printer tests → printer).

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
