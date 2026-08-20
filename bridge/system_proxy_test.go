package bridge

import (
	"strings"
	"testing"
)

func TestEnsureLoopbackBypass(t *testing.T) {
	cases := map[string][]string{
		"":                             {"localhost", "127.0.0.1", "::1"},
		"example.com":                  {"localhost", "127.0.0.1", "::1", "example.com"},
		"localhost;::1;foo":            {"localhost", "127.0.0.1", "::1", "foo"},
		"127.0.0.1;bar;::1;localhost":  {"localhost", "127.0.0.1", "::1", "bar"},
	}
	for input, want := range cases {
		got := strings.Split(ensureLoopbackBypass(input), ";")
		if len(got) != len(want) {
			t.Fatalf("ensureLoopbackBypass(%q) = %v, want %v", input, got, want)
		}
		wantSet := make(map[string]bool)
		for _, w := range want {
			wantSet[w] = true
		}
		for _, g := range got {
			if !wantSet[g] {
				t.Fatalf("ensureLoopbackBypass(%q) = %v, want %v", input, got, want)
			}
		}
	}
}
