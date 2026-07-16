package httpapi

import (
	"errors"
	"net/http"

	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/apps/auth-service/internal/auth"
	"github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/httpx"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  *auth.Service
	validate *validator.Validate
}

func NewHandler(service *auth.Service) *Handler { return &Handler{service, validator.New()} }

type registerInput struct {
	Name     string `json:"name" validate:"required,min=2,max=150"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}
type loginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type refreshInput struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in registerInput
	if httpx.DecodeJSON(r, &in) != nil || h.validate.Struct(in) != nil {
		httpx.Error(w, 400, "validation_error", "invalid payload")
		return
	}
	u, err := h.service.Register(r.Context(), in.Name, in.Email, in.Password)
	if err != nil {
		httpx.Error(w, 409, "user_conflict", "could not create user")
		return
	}
	httpx.JSON(w, 201, u)
}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	if httpx.DecodeJSON(r, &in) != nil || h.validate.Struct(in) != nil {
		httpx.Error(w, 400, "validation_error", "invalid payload")
		return
	}
	tokens, err := h.service.Login(r.Context(), in.Email, in.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		httpx.Error(w, 401, "invalid_credentials", "invalid credentials")
		return
	}
	if err != nil {
		httpx.Error(w, 500, "internal_error", "could not login")
		return
	}
	httpx.JSON(w, 200, tokens)
}
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in refreshInput
	if httpx.DecodeJSON(r, &in) != nil || h.validate.Struct(in) != nil {
		httpx.Error(w, 400, "validation_error", "invalid payload")
		return
	}
	tokens, err := h.service.Refresh(r.Context(), in.RefreshToken)
	if err != nil {
		httpx.Error(w, 401, "invalid_refresh", "invalid refresh token")
		return
	}
	httpx.JSON(w, 200, tokens)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var in refreshInput
	if httpx.DecodeJSON(r, &in) != nil {
		httpx.Error(w, 400, "validation_error", "invalid payload")
		return
	}
	_ = h.service.Logout(r.Context(), in.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}
