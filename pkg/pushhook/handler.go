package pushhook

import (
	"net/http"

	"github.com/go-logr/logr"
)

type Handler struct {
	log logr.Logger
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func NewHandler(logger logr.Logger) *Handler {
	return &Handler{log: logger}
}
