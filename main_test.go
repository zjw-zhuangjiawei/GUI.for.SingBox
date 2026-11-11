package main

import (
	"reflect"
	"testing"
)

func TestPayloadRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		args []any
	}{
		{"no args", nil},
		{"one string", []any{"text"}},
		{"one map", []any{map[string]any{"type": "done"}}},
		{"one slice", []any{[]string{"--data-dir=/tmp"}}},
		{"six args", []any{"req1", "GET", "/x", nil, "body", 1}},
	}
	for _, c := range cases {
		// v3 keeps a lone value as-is and turns multiple values into []any,
		// which is what normalizePayload/toArgs have to undo.
		sent := normalizePayload(c.args)
		var data any
		if len(sent) == 1 {
			data = sent[0]
		} else if len(sent) > 1 {
			data = sent
		}
		if got := toArgs(data); !reflect.DeepEqual(got, c.args) && !(len(got) == 0 && len(c.args) == 0) {
			t.Errorf("%s: toArgs(normalizePayload(%v)) = %v", c.name, c.args, got)
		}
	}
}
