## Inline Table
Inline tables provide a more compact syntax for expressing tables. They are
especially useful for grouped data that can otherwise quickly become verbose.
Inline tables are fully defined within curly braces: `{` and `}`. Within the
braces, zero or more comma-separated key/value pairs may appear. Key/value pairs
take the same form as key/value pairs in standard tables. All value types are
allowed, including inline tables.

Inline tables are intended to appear on a single line. A terminating comma (also
called trailing comma) is not permitted after the last key/value pair in an
inline table. No newlines are allowed between the curly braces unless they are
valid within a value. Even so, it is strongly discouraged to break an inline
table onto multiples lines. If you find yourself gripped with this desire, it
means you should be using standard tables.

```toml
name = { first = "Tom", last = "Preston-Werner" }
point = { x = 1, y = 2 }
animal = { type.name = "pug" }
```

The inline tables above are identical to the following standard table
definitions:

```toml
[name]
first = "Tom"
last = "Preston-Werner"

[point]
x = 1
y = 2

[animal]
type.name = "pug"
```

Inline tables are fully self-contained and define all keys and sub-tables within
them. Keys and sub-tables cannot be added outside the braces.

```toml
[product]
type = { name = "Nail" }
# type.edible = false  # INVALID
```

Similarly, inline tables cannot be used to add keys or sub-tables to an
already-defined table.

```toml
[product]
type.name = "Nail"
# type = { edible = false }  # INVALID
```
