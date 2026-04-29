## Table
Tables (also known as hash tables or dictionaries) are collections of key/value
pairs. They are defined by headers, with square brackets on a line by
themselves. You can tell headers apart from arrays because arrays are only ever
values.

```toml
[table]
```

Under that, and until the next header or EOF, are the key/values of that table.
Key/value pairs within tables are not guaranteed to be in any specific order.

```toml
[table-1]
key1 = "some string"
key2 = 123

[table-2]
key1 = "another string"
key2 = 456
```

Naming rules for tables are the same as for keys (see definition of
[Keys](#keys) above).

```toml
[dog."tater.man"]
type.name = "pug"
```

In JSON land, that would give you the following structure:

```json
{ "dog": { "tater.man": { "type": { "name": "pug" } } } }
```

Whitespace around the key is ignored. However, best practice is to not use any
extraneous whitespace.

```toml
[a.b.c]            # this is best practice
[ d.e.f ]          # same as [d.e.f]
[ g .  h  . i ]    # same as [g.h.i]
[ j . "ʞ" . 'l' ]  # same as [j."ʞ".'l']
```

Indentation is treated as whitespace and ignored.

You don't need to specify all the super-tables if you don't want to. TOML knows
how to do it for you.

```toml
# [x] you
# [x.y] don't
# [x.y.z] need these
[x.y.z.w] # for this to work

[x] # defining a super-table afterward is ok
```

Empty tables are allowed and simply have no key/value pairs within them.

Like keys, you cannot define a table more than once. Doing so is invalid.

```
# DO NOT DO THIS

[fruit]
apple = "red"

[fruit]
orange = "orange"
```

```
# DO NOT DO THIS EITHER

[fruit]
apple = "red"

[fruit.apple]
texture = "smooth"
```

Defining tables out-of-order is discouraged.

```toml
# VALID BUT DISCOURAGED
[fruit.apple]
[animal]
[fruit.orange]
```

```toml
# RECOMMENDED
[fruit.apple]
[fruit.orange]
[animal]
```

The top-level table, also called the root table, starts at the beginning of the
document and ends just before the first table header (or EOF). Unlike other
tables, it is nameless and cannot be relocated.

```toml
# Top-level table begins.
name = "Fido"
breed = "pug"

# Top-level table ends.
[owner]
name = "Regina Dogman"
member_since = 1999-08-04
```

Dotted keys create and define a table for each key part before the last one,
provided that such tables were not previously created.

```toml
fruit.apple.color = "red"
# Defines a table named fruit
# Defines a table named fruit.apple

fruit.apple.taste.sweet = true
# Defines a table named fruit.apple.taste
# fruit and fruit.apple were already created
```

Since tables cannot be defined more than once, redefining such tables using a
`[table]` header is not allowed. Likewise, using dotted keys to redefine tables
already defined in `[table]` form is not allowed. The `[table]` form can,
however, be used to define sub-tables within tables defined via dotted keys.

```toml
[fruit]
apple.color = "red"
apple.taste.sweet = true

# [fruit.apple]  # INVALID
# [fruit.apple.taste]  # INVALID

[fruit.apple.texture]  # you can add sub-tables
smooth = true
```
