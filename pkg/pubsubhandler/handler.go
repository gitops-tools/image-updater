package pubsubhandler

import (
	"context"

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

	hook, err := h.parser(m.Data())
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
