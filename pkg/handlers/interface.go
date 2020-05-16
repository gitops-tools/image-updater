package handlers

type nameGenerator interface {
	prefixedName(s string) string
}
