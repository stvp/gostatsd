package statsd

import (
	"testing"
)

func TestParseUrl(t *testing.T) {
	tests := []parseUrlTestcase{
		{"", "", "", false},
		{"x", "", "", false},
		{"foo.com", "", "", false},
		{"statsd://a.b.com", "a.b.com", "", true},
		{"statsd://a.b.com/foo.bar", "a.b.com", "foo.bar.", true},
		{"statsd://a.b.com/foo.bar.", "a.b.com", "foo.bar.", true},
	}

	for _, test := range tests {
		host, prefix, err := parseUrl(test.url)
		if test.good {
			if test.host != host {
				t.Errorf("Expected host %#v but got %#v", test.host, host)
			}
			if test.prefix != prefix {
				t.Errorf("Expected prefix %#v but got %#v", test.prefix, prefix)
			}
		} else {
			if err == nil {
				t.Errorf("parseUrl(%#v) should return an error", test.url)
			}
		}
	}
}
