# 4. Arrays

Source: RFC 8259, Section 5 ("Arrays").

## 4.1 Structure

An array is represented as square brackets surrounding zero or more values
(called **elements**). Elements are separated by commas.

```abnf
array = begin-array [ value *( value-separator value ) ] end-array
```

Allowed forms:

- Empty array: `[]`
- Single element: `[ v ]`
- Multiple elements: `[ v1, v2, v3 ]`

Not allowed by the grammar:

- Trailing comma: `[ v, ]`
- Leading comma / elision: `[ , v ]`, `[ v,, v ]`

## 4.2 Heterogeneous element types

There is no requirement that the values in an array be of the same type. An
array may mix any of the seven value kinds freely:

```
[ 1, "two", true, null, [3, 4], { "k": 5 } ]
```

Implementer guidance: a Go AST representation typically uses a slice of an
interface type (or a tagged-union value type) to hold mixed-type elements.
