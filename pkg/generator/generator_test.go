package generator

import (
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	g := New()

	t.Run("valid length", func(t *testing.T) {
		length := 10

		str, err := g.Generate(length)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(str) != length {
			t.Fatalf("expected length %d, got %d", length, len(str))
		}

		for _, char := range str {
			if !strings.ContainsRune(charset, char) {
				t.Fatalf("character %c is not in allowed charset", char)
			}
		}
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := g.Generate(0)
		if err == nil {
			t.Fatal("expected error for zero length")
		}
	})
}
