package gcr

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gitops-tools/image-updater/pkg/hooks"
)

// PushMessage is a struct for the GCR push event
type PushMessage struct {
	Action string `json:"action,omitempty"`
	Digest string `json:"digest,omitempty"`
	Tag    string `json:"tag,omitempty"`
}

// PushedImageURL is an implementation of the hooks.PushEvent interface.
func (m PushMessage) PushedImageURL() string {
	return m.Tag
}

// EventRepository is an implementation of the hooks.PushEvent interface.
func (m PushMessage) EventRepository() string {
	return strings.Split(m.Tag, ":")[0]
}

// EventTag is an implementation of the hooks.PushEvent interface.
func (m PushMessage) EventTag() string {
	return strings.Split(m.Tag, ":")[1]
}

// Parse parses a payload into a GCR PushEvent
func Parse(payload []byte) (hooks.PushEvent, error) {
	msg := &PushMessage{}

	err := json.Unmarshal(payload, &msg)
	if err != nil {
		return nil, err
	}

	if msg.Tag == "" {
		return nil, errors.New("tag is empty")
	}

	return msg, nil
}
