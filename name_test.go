package statsd

import "testing"

func TestJoin(t *testing.T) {
	tests := []struct {
		parts  []string
		expect string
	}{
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "a.b.c"},
		{[]string{"a", " über.mensch!"}, "a.uber_mensch"},
	}

	for i, test := range tests {
		got := Join(test.parts...)
		if got != test.expect {
			t.Errorf("tests[%d]: expected: %#v, got: %#v", i, test.expect, got)
		}
	}
}
