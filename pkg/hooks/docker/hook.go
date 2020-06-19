package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gitops-tools/image-hooks/pkg/hooks"
)

// Parse takes an http.Request and parses it into a Docker webhook event.
func Parse(req *http.Request) (hooks.PushEvent, error) {
	// TODO: LimitReader
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	h := &Webhook{}
	err = json.Unmarshal(data, h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// Webhook is a struct for the Docker Hub webhook event.
type Webhook struct {
	CallbackURL string      `json:"callback_url"`
	PushData    *PushData   `json:"push_data"`
	Repository  *Repository `json:"repository"`
}

// PushedImageURL is an implementation of the hooks.PushEvent interface.
func (p Webhook) PushedImageURL() string {
	return fmt.Sprintf("%s:%s", p.Repository.RepoName, p.PushData.Tag)
}

// EventRepository is an implementation of the hooks.PushEvent interface.
func (p Webhook) EventRepository() string {
	return p.Repository.RepoName
}

// PushData is part of the Webhook struct.
type PushData struct {
	Images   []string `json:"images"`
	PushedAt float64  `json:"pushed_at"`
	Pusher   string   `json:"pusher"`
	Tag      string   `json:"tag"`
}

// Repository is part of the Webhook struct.
type Repository struct {
	RepoName        string  `json:"repo_name"`
	Name            string  `json:"name"`
	Namespace       string  `json:"namespace"`
	Owner           string  `json:"owner"`
	Description     string  `json:"description"`
	FullDescription string  `json:"full_description"`
	RepoURL         string  `json:"repo_url"`
	Dockerfile      string  `json:"dockerfile"`
	Status          string  `json:"status"`
	IsOfficial      bool    `json:"is_official"`
	IsPrivate       bool    `json:"is_private"`
	IsTrusted       bool    `json:"is_trusted"`
	DateCreated     float64 `json:"date_created"`
	StarCount       int64   `json:"star_count"`
	CommentCount    int64   `json:"comment_count"`
}
