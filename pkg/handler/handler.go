package handler

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/gitops-tools/image-updater/pkg/hooks"
	"github.com/gitops-tools/image-updater/pkg/updater"
)

// Handler parses and processes hook notifications.
type Handler struct {
	log     *zap.SugaredLogger
	updater *updater.Updater
	parser  hooks.PushEventParser
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Infow("processing hook request")
	hook, err := h.parser(r)
	if err != nil {
		h.log.Errorf("failed to parse request %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.updater.UpdateFromHook(r.Context(), hook)

	if err != nil {
		h.log.Errorf("hook update failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// New creates and returns a new Handler.
func New(logger *zap.SugaredLogger, u *updater.Updater, p hooks.PushEventParser) *Handler {
	return &Handler{log: logger, updater: u, parser: p}
}
