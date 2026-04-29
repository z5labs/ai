## Array
Arrays are square brackets with values inside. Whitespace is ignored. Elements
are separated by commas. Arrays can contain values of the same data types as
allowed in key/value pairs. Values of different types may be mixed.

```toml
integers = [ 1, 2, 3 ]
colors = [ "red", "yellow", "green" ]
nested_arrays_of_ints = [ [ 1, 2 ], [3, 4, 5] ]
nested_mixed_array = [ [ 1, 2 ], ["a", "b", "c"] ]
string_array = [ "all", 'strings', """are the same""", '''type''' ]

# Mixed-type arrays are allowed
numbers = [ 0.1, 0.2, 0.5, 1, 2, 5 ]
contributors = [
  "Foo Bar <foo@example.com>",
  { name = "Baz Qux", email = "bazqux@example.com", url = "https://example.com/bazqux" }
]
```

Arrays can span multiple lines. A terminating comma (also called a trailing
comma) is permitted after the last value of the array. Any number of newlines
and comments may precede values, commas, and the closing bracket. Indentation
between array values and commas is treated as whitespace and ignored.

```toml
integers2 = [
  1, 2, 3
]

integers3 = [
  1,
  2, # this is ok
]
```
