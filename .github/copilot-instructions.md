# Copilot Instructions

This repository contains Claude Code agents and skills maintained by z5labs. It holds markdown-defined agent definitions and skill templates—not traditional application code.

## Repository Structure

- `agents/` — Agent definitions (markdown files with YAML frontmatter)
- `skills/` — Skill definitions, each in a subdirectory with a `SKILL.md` file

## Key Domain: File Library Packages

The agents and skills here support **file library packages**—Go packages that follow a **Tokenizer → Parser → AST → Printer** pipeline for parsing and formatting file formats.

### Core Components

1. **Tokenizer**: Converts source text into tokens via `iter.Seq2[Token, error]`
2. **Parser**: Converts tokens into an AST using `iter.Pull2()` for pull-based consumption
3. **Printer**: Formats AST back to source text

### Implementation Workflow (Test-First)

When implementing features in file library packages, follow this strict order:

1. Tokenizer tests → Tokenizer implementation
2. Parser tests → Parser implementation
3. Printer tests → Printer implementation

Run `go test -race ./...` after each step.

## Go Conventions

- State machine pattern with recursive action functions for tokenizer, parser, and printer
- `iter.Seq2` / `iter.Pull2` for token streaming
- Table-driven tests with `t.Parallel()` and `testify/require`
- Complex types must use the inner action loop pattern (no inline for-loops)

## Commit Convention

Format: `scope: description` (e.g., `ai: implement file library agent`). Use lowercase descriptions.
