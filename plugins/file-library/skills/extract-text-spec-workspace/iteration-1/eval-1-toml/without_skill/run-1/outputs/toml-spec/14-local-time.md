## Local Time
If you include only the time portion of an [RFC
3339](https://tools.ietf.org/html/rfc3339) formatted date-time, it will
represent that time of day without any relation to a specific day or any offset
or timezone.

```toml
lt1 = 07:32:00
lt2 = 00:32:00.999999
```

Millisecond precision is required. Further precision of fractional seconds is
implementation-specific. If the value contains greater precision than the
implementation can support, the additional precision must be truncated, not
rounded.
