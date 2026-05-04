# PropAtom

`PROP` (sub-atom of `HEAD`). Carries arbitrary metadata as name/value string
pairs.

## Encoding

`PROP` is a [string table](string-table-atom.md) with an **even** number of
strings. Strings are paired (name, value, name, value, …); pair `i`'s name
is string `2i` and its value is string `2i+1`.

Reserved name prefixes:

- `sim/` — public X-Plane properties.
- `laminar/` — Laminar Research private use.
- Any other prefix is open to vendors; convention is to prefix with the
  organisation name.

Defined `sim/*` properties:

| Property | Default if missing | Definition |
|---|---|---|
| `sim/west`           | (required) | Western edge in degrees longitude |
| `sim/east`           | (required) | Eastern edge in degrees longitude |
| `sim/south`          | (required) | Southern edge in degrees latitude (see ambiguity) |
| `sim/north`          | (required) | Northern edge in degrees latitude (see ambiguity) |
| `sim/planet`         | `earth`    | One of `earth` or `mars` |
| `sim/creation_agent` | (blank)    | Tool that wrote the file |
| `sim/author`         | (blank)    | Author of the file |
| `sim/require_object` | n/a        | Force-draw rules; value is `<rendering-level>/<first-defn-index>`; may repeat |
| `sim/require_facade` | n/a        | Same shape as `sim/require_object`, applied to facades |

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `StringTableAtom` | Pairs | Even number of strings; consecutive pairs are name/value. |

## Ambiguities

> **Ambiguity:** The spec text describes `sim/south` as "the northern edge of
> the DSF file in degrees latitude" and `sim/north` as "the southern edge".
> This is a documentation typo: by name they are obviously the south/north
> edges respectively, and X-Plane's writer treats them that way. Decoders
> should treat `sim/south` as the southern edge and `sim/north` as the
> northern edge.

> **Ambiguity:** The spec does not say whether `PROP` must come first inside
> `HEAD`. Convention is that `PROP` is the only sub-atom today, so there's
> no ordering question in practice.
