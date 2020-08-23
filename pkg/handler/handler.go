package handler

import (
	"net/http"

	"github.com/go-logr/logr"

	"github.com/gitops-tools/image-updater/pkg/applier"
	"github.com/gitops-tools/image-updater/pkg/hooks"
)

// Handler parses and processes hook notifications.
type Handler struct {
	log     logr.Logger
	applier *applier.Applier
	parser  hooks.PushEventParser
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("processing hook request")
	hook, err := h.parser(r)
	if err != nil {
		h.log.Error(err, "failed to parse request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.applier.UpdateFromHook(r.Context(), hook)

	if err != nil {
		h.log.Error(err, "hook update failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// New creates and returns a new Handler.
func New(logger logr.Logger, u *applier.Applier, p hooks.PushEventParser) *Handler {
	return &Handler{log: logger, applier: u, parser: p}
}
