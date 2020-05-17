package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bigkevmcd/image-hooks/pkg/hooks"
)

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

type Webhook struct {
	CallbackURL string      `json:"callback_url"`
	PushData    *PushData   `json:"push_data"`
	Repository  *Repository `json:"repository"`
}

func (p Webhook) PushedImageURL() string {
	return fmt.Sprintf("%s:%s", p.Repository.RepoName, p.PushData.Tag)
}

func (p Webhook) EventRepository() string {
	return p.Repository.RepoName
}

type PushData struct {
	Images   []string `json:"images"`
	PushedAt float64  `json:"pushed_at"`
	Pusher   string   `json:"pusher"`
	Tag      string   `json:"tag"`
}

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
	StarCount       int64   `star_count"`
	CommentCount    int64   `json:"comment_count"`
}
