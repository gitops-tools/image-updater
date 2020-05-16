package handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/bigkevmcd/image-hooks/pkg/hooks/quay"
)

type Handler struct {
	log     *zap.SugaredLogger
	updater *Updater
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Infow("processing hook request")
	hook, err := quay.ParseRepositoryPush(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.updater.Update(r.Context(), hook)
	if err != nil {
		h.log.Errorf("hook update failed: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func NewHandler(logger *zap.SugaredLogger, u *Updater) *Handler {
	return &Handler{log: logger, updater: u}
}
