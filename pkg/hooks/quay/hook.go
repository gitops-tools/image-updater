package quay

import (
	"encoding/json"
	"fmt"

	"github.com/gitops-tools/image-updater/pkg/hooks"
)

// Parse parses a payload it into a Quay.io Push hook if possible.
func Parse(payload []byte) (hooks.PushEvent, error) {
	h := &RepositoryPushHook{}
	err := json.Unmarshal(payload, h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// RepositoryPushHook is a struct for the Quay.io push event.
type RepositoryPushHook struct {
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Namespace   string   `json:"namespace"`
	DockerURL   string   `json:"docker_url"`
	Homepage    string   `json:"homepage"`
	UpdatedTags []string `json:"updated_tags,omitempty"`
}

// PushedImageURL is an implementation of the hooks.PushEvent interface.
func (p RepositoryPushHook) PushedImageURL() string {
	return fmt.Sprintf("%s:%s", p.DockerURL, p.UpdatedTags[0])
}

// EventRepository is an implementation of the hooks.PushEvent interface.
func (p RepositoryPushHook) EventRepository() string {
	return p.Repository
}

// EventTag is an implementation of the hooks.PushEvent interface.
func (p RepositoryPushHook) EventTag() string {
	return p.UpdatedTags[0]
}
