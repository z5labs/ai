## Key/Value Pair
The primary building block of a TOML document is the key/value pair.

Keys are on the left of the equals sign and values are on the right. Whitespace
is ignored around key names and values. The key, equals sign, and value must be
on the same line (though some values can be broken over multiple lines).

```toml
key = "value"
```

Values must have one of the following types.

- [String](#string)
- [Integer](#integer)
- [Float](#float)
- [Boolean](#boolean)
- [Offset Date-Time](#offset-date-time)
- [Local Date-Time](#local-date-time)
- [Local Date](#local-date)
- [Local Time](#local-time)
- [Array](#array)
- [Inline Table](#inline-table)

Unspecified values are invalid.

```toml
key = # INVALID
```

There must be a newline (or EOF) after a key/value pair. (See [Inline
Table](#inline-table) for exceptions.)

```
first = "Tom" last = "Preston-Werner" # INVALID
```
