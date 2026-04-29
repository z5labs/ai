## Offset Date-Time
To unambiguously represent a specific instant in time, you may use an [RFC
3339](https://tools.ietf.org/html/rfc3339) formatted date-time with offset.

```toml
odt1 = 1979-05-27T07:32:00Z
odt2 = 1979-05-27T00:32:00-07:00
odt3 = 1979-05-27T00:32:00.999999-07:00
```

For the sake of readability, you may replace the T delimiter between date and
time with a space character (as permitted by RFC 3339 section 5.6).

```toml
odt4 = 1979-05-27 07:32:00Z
```

Millisecond precision is required. Further precision of fractional seconds is
implementation-specific. If the value contains greater precision than the
implementation can support, the additional precision must be truncated, not
rounded.
