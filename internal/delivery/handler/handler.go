package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"shortener/internal/domain"
)

type URLUseCase interface {
	Shorten(ctx context.Context, original string) (string, error)
	GetOriginal(ctx context.Context, short string) (string, error)
}

type Handler struct {
	uc  URLUseCase
	log *slog.Logger
}

func NewHandler(uc URLUseCase, log *slog.Logger) *Handler {
	return &Handler{
		uc:  uc,
		log: log,
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ShortenRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		h.respondError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if strings.TrimSpace(req.URL) == "" {
		h.respondError(w, http.StatusBadRequest, "url is required")
		return
	}

	short, err := h.uc.Shorten(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		h.log.Error(
			"failed to shorten url",
			slog.String("url", req.URL),
			slog.String("error", err.Error()),
		)
		h.respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondJSON(w, http.StatusCreated, ShortenResponse{ShortURL: short})
}

func (h *Handler) GetOriginal(w http.ResponseWriter, r *http.Request) {
	short := r.PathValue("short")
	if short == "" {
		h.respondError(w, http.StatusBadRequest, "short value is required")
		return
	}

	orig, err := h.uc.GetOriginal(r.Context(), short)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidShort):
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, domain.ErrNotFound):
			h.respondError(w, http.StatusNotFound, "url not found")
			return
		default:
			h.log.Error(
				"failed to get original url",
				slog.String("short", short),
				slog.String("error", err.Error()),
			)
			h.respondError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	h.respondJSON(w, http.StatusOK, OriginalResponse{OriginalURL: orig})
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{Error: message})
}
