package cmd

import "testing"

func TestMaskEmail(t *testing.T) {
	cases := map[string]string{
		"":                   "",
		"me@gmail.com":       "m***@gmail.com",
		"mahendra@gmail.com": "ma***@gmail.com",
		"a@b.io":             "a***@b.io",
		"nodomain":           "no***",
	}
	for in, want := range cases {
		if got := maskEmail(in); got != want {
			t.Errorf("maskEmail(%q) = %q, want %q", in, got, want)
		}
	}
}
