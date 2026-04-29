## Local Date-Time
If you omit the offset from an [RFC 3339](https://tools.ietf.org/html/rfc3339)
formatted date-time, it will represent the given date-time without any relation
to an offset or timezone. It cannot be converted to an instant in time without
additional information. Conversion to an instant, if required, is
implementation-specific.

```toml
ldt1 = 1979-05-27T07:32:00
ldt2 = 1979-05-27T00:32:00.999999
```

Millisecond precision is required. Further precision of fractional seconds is
implementation-specific. If the value contains greater precision than the
implementation can support, the additional precision must be truncated, not
rounded.
