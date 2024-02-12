package parser

import "testing"

func FuzzParseSrc(f *testing.F) {
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ParseSrc(s)
	})
}
