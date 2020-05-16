package handlers

import (
	"math/rand"
	"testing"
)

func TestGenerator(t *testing.T) {
	g := randomNameGenerator{rand: rand.New(rand.NewSource(100))}

	name := g.prefixedName("testing-")

	if name != "testing-DlPsU" {
		t.Fatalf("got %v, want %v", name, "testing-DlPsU")
	}
}
