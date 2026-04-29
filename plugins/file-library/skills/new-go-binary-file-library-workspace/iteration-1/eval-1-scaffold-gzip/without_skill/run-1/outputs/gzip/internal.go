package gzip

import "fmt"

// errNotImplemented is returned by scaffolded methods that have not yet been
// filled in. It is intentionally unexported so callers cannot rely on it.
func errNotImplemented(name string) error {
	return fmt.Errorf("gzip: %s not implemented", name)
}
