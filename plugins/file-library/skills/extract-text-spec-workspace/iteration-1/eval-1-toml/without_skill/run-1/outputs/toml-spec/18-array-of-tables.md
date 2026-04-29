## Array of Tables
The last syntax that has not yet been described allows writing arrays of tables.
These can be expressed by using a header with a name in double brackets. The
first instance of that header defines the array and its first table element, and
each subsequent instance creates and defines a new table element in that array.
The tables are inserted into the array in the order encountered.

```toml
[[products]]
name = "Hammer"
sku = 738594937

[[products]]  # empty table within the array

[[products]]
name = "Nail"
sku = 284758393

color = "gray"
```

In JSON land, that would give you the following structure.

```json
{
  "products": [
    { "name": "Hammer", "sku": 738594937 },
    { },
    { "name": "Nail", "sku": 284758393, "color": "gray" }
  ]
}
```

Any reference to an array of tables points to the most recently defined table
element of the array. This allows you to define sub-tables, and even sub-arrays
of tables, inside the most recent table.

```toml
[[fruits]]
name = "apple"

[fruits.physical]  # subtable
color = "red"
shape = "round"

[[fruits.varieties]]  # nested array of tables
name = "red delicious"

[[fruits.varieties]]
name = "granny smith"


[[fruits]]
name = "banana"

[[fruits.varieties]]
name = "plantain"
```

The above TOML maps to the following JSON.

```json
{
  "fruits": [
    {
      "name": "apple",
      "physical": {
        "color": "red",
        "shape": "round"
      },
      "varieties": [
        { "name": "red delicious" },
        { "name": "granny smith" }
      ]
    },
    {
      "name": "banana",
      "varieties": [
        { "name": "plantain" }
      ]
    }
  ]
}
```

If the parent of a table or array of tables is an array element, that element
must already have been defined before the child can be defined. Attempts to
reverse that ordering must produce an error at parse time.

```
# INVALID TOML DOC
[fruit.physical]  # subtable, but to which parent element should it belong?
color = "red"
shape = "round"

[[fruit]]  # parser must throw an error upon discovering that "fruit" is
           # an array rather than a table
name = "apple"
```

Attempting to append to a statically defined array, even if that array is empty,
must produce an error at parse time.

```
# INVALID TOML DOC
fruits = []

[[fruits]] # Not allowed
```

Attempting to define a normal table with the same name as an already established
array must produce an error at parse time. Attempting to redefine a normal table
as an array must likewise produce a parse-time error.

```
# INVALID TOML DOC
[[fruits]]
name = "apple"

[[fruits.varieties]]
name = "red delicious"

# INVALID: This table conflicts with the previous array of tables
[fruits.varieties]
name = "granny smith"

[fruits.physical]
color = "red"
shape = "round"

# INVALID: This array of tables conflicts with the previous table
[[fruits.physical]]
color = "green"
```

You may also use inline tables where appropriate:

```toml
points = [ { x = 1, y = 2, z = 3 },
           { x = 7, y = 8, z = 9 },
           { x = 2, y = 4, z = 8 } ]
```
