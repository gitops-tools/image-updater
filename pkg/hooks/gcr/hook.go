package gcr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// Parse takes an http request and returns a PushEvent
func Parse(req *http.Request) (hooks.PushEvent, error) {
	data, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}
	msg := &PushMessage{}

	err = json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}

	if msg.Tag == "" {
		return nil, fmt.Errorf("Tag is empty")
	}

	return msg, nil
}
