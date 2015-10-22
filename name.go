package statsd

import (
	"strings"

	"github.com/stvp/slug"
)

// Split takes a valid Statsd metric name and returns it's constituent parts.
func Split(name string) (parts []string) {
	parts = []string{}
	for _, part := range strings.Split(name, ".") {
		if len(part) > 0 {
			parts = append(parts, part)
		}
	}
	return
}

// Join takes a slice of strings (a metric "path") and returns a valid Statsd
// metric name. It cleans all parts by slug-ifying them.
func Join(parts []string) (name string) {
	clean := make([]string, len(parts))
	for i := 0; i < len(parts); i++ {
		clean[i] = slug.Clean(parts[i])
	}
	return strings.Join(clean, ".")
}
