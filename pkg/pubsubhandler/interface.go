package pubsubhandler

type message interface {
	Ack()
	Data() []byte
}
