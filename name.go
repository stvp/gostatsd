package statsd

import (
	"strings"

	"github.com/stvp/slug"
)

// Join takes a slice of strings (a metric "path") and returns a valid Statsd
// metric name. It cleans all parts by slug-ifying them.
func Join(parts []string) (name string) {
	clean := make([]string, len(parts))
	for i := 0; i < len(parts); i++ {
		clean[i] = slug.Clean(parts[i])
	}
	return strings.Join(clean, ".")
}
