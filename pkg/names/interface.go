package names

// Generator is implemented by values that generate a prefixed-nane.
type Generator interface {
	PrefixedName(s string) string
}
