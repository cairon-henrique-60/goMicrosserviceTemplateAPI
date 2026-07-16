package order

import (
	"errors"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type Handler struct {
	s *Service
	v *validator.Validate
}

func NewHandler(s *Service) *Handler { return &Handler{s, validator.New()} }
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if httpx.DecodeJSON(r, &in) != nil || h.v.Struct(in) != nil {
		httpx.Error(w, 400, "validation_error", "invalid payload")
		return
	}
	o, err := h.s.Create(r.Context(), in)
	if err != nil {
		httpx.Error(w, 422, "order_failed", err.Error())
		return
	}
	httpx.JSON(w, 201, o)
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.s.List(r.Context())
	if err != nil {
		httpx.Error(w, 500, "list_failed", "could not list orders")
		return
	}
	httpx.JSON(w, 200, items)
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	o, err := h.s.Get(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, 404, "not_found", "order not found")
		return
	}
	if err != nil {
		httpx.Error(w, 500, "get_failed", "could not get order")
		return
	}
	httpx.JSON(w, 200, o)
}
