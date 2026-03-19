package handler

import "net/http"

func NewRouter(urlHandler *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/shorten", urlHandler.ShortenURL)
	mux.HandleFunc("GET /api/original/{short}", urlHandler.GetOriginal)

	return mux
}
