package pubsubhandler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gitops-tools/image-updater/pkg/applier"
	"github.com/gitops-tools/image-updater/pkg/hooks"
	"github.com/go-logr/logr"
)

// Handler parses and processes pubsub messages.
type Handler struct {
	applier *applier.Applier
	log     logr.Logger
	parser  hooks.PushEventParser
}

// New creates and returns a new Handler.
func New(logger logr.Logger, u *applier.Applier, p hooks.PushEventParser) *Handler {
	return &Handler{log: logger, applier: u, parser: p}
}

// Handle acks, parses and processes pubsub messages
func (h *Handler) Handle(ctx context.Context, m message) {
	h.log.Info("processing hook request")

	req, err := h.convertToHTTP(ctx, m)
	if err != nil {
		h.log.Error(err, "failed to parse message")
		return
	}
	hook, err := h.parser(req)
	if err != nil {
		h.log.Error(err, "failed to parse request")
		return
	}

	err = h.applier.UpdateFromHook(ctx, hook)

	if err != nil {
		h.log.Error(err, "hook update failed")
		return
	}

	m.Ack()
}

func (h *Handler) convertToHTTP(ctx context.Context, m message) (*http.Request, error) {
	r := strings.NewReader(string(m.Data()))

	req, err := http.NewRequest("POST", "application/json", r)
	if err != nil {
		return nil, err
	}
	return req, nil
}
