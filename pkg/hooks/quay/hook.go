package quay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bigkevmcd/image-hooks/pkg/hooks"
)

func ParseRequest(req *http.Request) (hooks.PushEvent, error) {
	// TODO: LimitReader
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	h := &RepositoryPushHook{}
	err = json.Unmarshal(data, h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

type RepositoryPushHook struct {
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Namespace   string   `json:"namespace"`
	DockerURL   string   `json:"docker_url"`
	Homepage    string   `json:"homepage"`
	UpdatedTags []string `json:"updated_tags,omitempty"`
}

func (p RepositoryPushHook) PushedImageURL() string {
	return fmt.Sprintf("%s:%s", p.DockerURL, p.UpdatedTags[0])
}

func (p RepositoryPushHook) EventRepository() string {
	return p.Repository
}
