## Comment
A hash symbol marks the rest of the line as a comment, except when inside a
string.

```toml
# This is a full-line comment
key = "value"  # This is a comment at the end of a line
another = "# This is not a comment"
```

Control characters other than tab (U+0000 to U+0008, U+000A to U+001F, U+007F)
are not permitted in comments.
