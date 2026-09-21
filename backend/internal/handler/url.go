package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"url-shortener/internal/models"
	"url-shortener/internal/service"

	"github.com/gorilla/mux"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{service: service}
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.service.ShortenURL(req.LongURL)
	if err != nil {
		switch err {
		case service.ErrInvalidURL:
			writeError(w, err.Error(), http.StatusBadRequest)
		default:
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, resp, http.StatusCreated)
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	longURL, err := h.service.RedirectURL(code)
	if err != nil {
		writeError(w, "URL not found", http.StatusNotFound)
		return
	}

	// Log analytics asynchronously
	referrer := r.Header.Get("Referer")
	userAgent := r.Header.Get("User-Agent")
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	}

	go h.service.LogClick(code, referrer, userAgent, ipAddress)

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	stats, entries, err := h.service.GetStats(code)
	if err != nil {
		writeError(w, "URL not found", http.StatusNotFound)
		return
	}

 response := map[string]interface{}{
		"stats":    stats,
		"analytics": entries,
	}

	writeJSON(w, response, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, models.ErrorResponse{Error: message}, status)
}
