package hook

import (
	"fmt"
	"math/rand"
	"time"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Generates a name from the prefix with an additional 5 random alphabetic
// characters.
// TODO: this should limit the length based on the prefix because branch names
// have a limit.
func randomNameGenerator(prefix string) string {
	charset := "abcdefghijklmnopqrstuvwyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s%s", prefix, b)
}
