package hook

type nameGenerator interface {
	prefixedName(s string) string
}
