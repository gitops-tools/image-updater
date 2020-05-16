package hooks

type PushEvent interface {
	PushedImageURL() string
}
