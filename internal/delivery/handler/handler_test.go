package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"shortener/internal/domain"
)

type useCaseMock struct {
	shortenFn     func(ctx context.Context, original string) (string, error)
	getOriginalFn func(ctx context.Context, short string) (string, error)
}

func (m *useCaseMock) Shorten(ctx context.Context, original string) (string, error) {
	return m.shortenFn(ctx, original)
}

func (m *useCaseMock) GetOriginal(ctx context.Context, short string) (string, error) {
	return m.getOriginalFn(ctx, short)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandler_ShortenURL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := NewHandler(&useCaseMock{
			shortenFn: func(ctx context.Context, original string) (string, error) {
				return "gbB7x9Z_ao", nil
			},
		}, testLogger())

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://alpha.dev"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.ShortenURL(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, `"short_url":"gbB7x9Z_ao"`) {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		h := NewHandler(&useCaseMock{}, testLogger())

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.ShortenURL(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		h := NewHandler(&useCaseMock{}, testLogger())

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://alpha.dev","extra":1}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.ShortenURL(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("json tail", func(t *testing.T) {
		h := NewHandler(&useCaseMock{}, testLogger())

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://alpha.dev"}{"x":1}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.ShortenURL(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("empty url", func(t *testing.T) {
		h := NewHandler(&useCaseMock{}, testLogger())

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"   "}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.ShortenURL(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		h := NewHandler(&useCaseMock{
			shortenFn: func(ctx context.Context, original string) (string, error) {
				return "", domain.ErrInvalidURL
			},
		}, testLogger())

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"invalid-url"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.ShortenURL(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		h := NewHandler(&useCaseMock{
			shortenFn: func(ctx context.Context, original string) (string, error) {
				return "", errors.New("boom")
			},
		}, testLogger())

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://alpha.dev"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.ShortenURL(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
	})
}

func TestHandler_GetOriginal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := NewHandler(&useCaseMock{
			getOriginalFn: func(ctx context.Context, short string) (string, error) {
				return "https://alpha.dev/page", nil
			},
		}, testLogger())

		req := httptest.NewRequest(http.MethodGet, "/api/original/gbB7x9Z_ao", nil)
		req.SetPathValue("short", "gbB7x9Z_ao")
		rr := httptest.NewRecorder()

		h.GetOriginal(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, `"original_url":"https://alpha.dev/page"`) {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("missing short path value", func(t *testing.T) {
		h := NewHandler(&useCaseMock{}, testLogger())

		req := httptest.NewRequest(http.MethodGet, "/api/original/", nil)
		rr := httptest.NewRecorder()

		h.GetOriginal(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("invalid short", func(t *testing.T) {
		h := NewHandler(&useCaseMock{
			getOriginalFn: func(ctx context.Context, short string) (string, error) {
				return "", domain.ErrInvalidShort
			},
		}, testLogger())

		req := httptest.NewRequest(http.MethodGet, "/api/original/abc", nil)
		req.SetPathValue("short", "abc")
		rr := httptest.NewRecorder()

		h.GetOriginal(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewHandler(&useCaseMock{
			getOriginalFn: func(ctx context.Context, short string) (string, error) {
				return "", domain.ErrNotFound
			},
		}, testLogger())

		req := httptest.NewRequest(http.MethodGet, "/api/original/gbB7x9Z_ao", nil)
		req.SetPathValue("short", "gbB7x9Z_ao")
		rr := httptest.NewRecorder()

		h.GetOriginal(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		h := NewHandler(&useCaseMock{
			getOriginalFn: func(ctx context.Context, short string) (string, error) {
				return "", errors.New("boom")
			},
		}, testLogger())

		req := httptest.NewRequest(http.MethodGet, "/api/original/gbB7x9Z_ao", nil)
		req.SetPathValue("short", "gbB7x9Z_ao")
		rr := httptest.NewRecorder()

		h.GetOriginal(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
	})
}
