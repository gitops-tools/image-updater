package names

import (
	"math/rand"
	"testing"
)

func TestGenerator(t *testing.T) {
	g := RandomGenerator{rand: rand.New(rand.NewSource(100))}

	name := g.PrefixedName("testing-")

	if name != "testing-DlPsU" {
		t.Fatalf("got %v, want %v", name, "testing-DlPsU")
	}
}
