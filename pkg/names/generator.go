package names

import (
	"fmt"
	"math/rand"
)

type RandomGenerator struct {
	rand *rand.Rand
}

func New(r *rand.Rand) *RandomGenerator {
	return &RandomGenerator{rand: r}
}

// Generates a name from the prefix with an additional 5 random alphabetic
// characters.
// TODO: this should limit the length based on the prefix because branch names
// have a limit.
func (g RandomGenerator) PrefixedName(prefix string) string {
	charset := "abcdefghijklmnopqrstuvwyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[g.rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s%s", prefix, b)
}
