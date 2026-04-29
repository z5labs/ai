// Copyright (c) 2026 z5labs
//
// Licensed under the MIT License (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT

// Package toml provides a tokenizer, parser, and printer for the TOML
// configuration file format.
//
// The package is organized around a Tokenizer -> Parser -> AST -> Printer
// pipeline:
//
//   - Tokenize streams [Token] values from an [io.Reader] using
//     [iter.Seq2].
//   - Parse consumes that token stream and produces a [*File] AST.
//   - Print walks an AST and writes its textual representation to an
//     [io.Writer].
package toml
