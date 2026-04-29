# 5. Numbers

Source: RFC 8259, Section 6 ("Numbers").

## 5.1 Structure

A JSON number is base 10, decimal digits only. It has an integer component
that may be prefixed with an optional minus sign, optionally followed by a
fraction part and/or an exponent part. Leading zeros are not allowed.

- A **fraction part** is a decimal point followed by one or more digits.
- An **exponent part** begins with `E` (uppercase or lowercase), which may
  be followed by a `+` or `-` sign. The `E` and optional sign are followed
  by one or more digits.
- Numeric values that cannot be represented by the grammar (such as
  `Infinity` and `NaN`) are not permitted.

```abnf
number = [ minus ] int [ frac ] [ exp ]

decimal-point = %x2E       ; .
digit1-9      = %x31-39    ; 1-9
e             = %x65 / %x45 ; e E
exp           = e [ minus / plus ] 1*DIGIT
frac          = decimal-point 1*DIGIT
int           = zero / ( digit1-9 *DIGIT )
minus         = %x2D       ; -
plus          = %x2B       ; +
zero          = %x30       ; 0
```

`DIGIT` is the RFC 5234 core rule: `%x30-39` (`0`-`9`).

## 5.2 What's valid

| Input        | Valid? | Reason                                             |
|--------------|--------|----------------------------------------------------|
| `0`          | yes    | bare zero                                          |
| `-0`         | yes    | minus followed by zero                             |
| `42`         | yes    | digit1-9 then digits                               |
| `-1.5`       | yes    | minus + int + frac                                 |
| `3.14e10`    | yes    | int + frac + exp                                   |
| `1E-9`       | yes    | int + uppercase E + minus + digit                  |
| `1e+0`       | yes    | int + lowercase e + plus + digit                   |
| `01`         | no     | leading zero on a multi-digit int                  |
| `.5`         | no     | int part is required                               |
| `5.`         | no     | frac requires at least one digit after the point   |
| `5e`         | no     | exp requires at least one digit                    |
| `+5`         | no     | leading plus is not in the `number` rule           |
| `0x1F`       | no     | hex not permitted                                  |
| `Infinity`   | no     | not in grammar                                     |
| `NaN`        | no     | not in grammar                                     |
| `1_000`      | no     | digit separators not permitted                     |

Tokenizer note: a number token ends as soon as the next byte cannot extend
the production. Numbers are not surrounded by `ws` in the grammar (they are
delimited only by structural characters or by reaching the trailing `ws` of
`JSON-text`).

## 5.3 Range and precision

The specification allows implementations to set limits on the range and
precision of numbers accepted.

Practical guidance from the RFC:

- IEEE 754 binary64 (double precision) is widely available; receivers that
  expect no more precision or range than this will be broadly interoperable.
  They will approximate JSON numbers to within that precision.
- Numbers like `1E400` or `3.141592653589793238462643383279` may indicate
  interoperability problems, since they exceed binary64 capacity.
- Integers in the range `[-(2**53)+1, (2**53)-1]` are interoperable in the
  sense that implementations will agree exactly on their numeric values.

Implementer guidance for Go: the AST may carry numbers as raw text plus an
optional decoded `float64` (or a `*big.Int` / `*big.Float` for arbitrary
precision). Preserving the original text is recommended for round-trip
printing because `float64` decoding is lossy.
