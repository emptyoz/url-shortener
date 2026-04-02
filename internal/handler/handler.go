package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/Vadim-Makhnev/url-shortener/internal/domain"
	"github.com/Vadim-Makhnev/url-shortener/internal/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type URLService interface {
	ShortenURL(originalURL string) (*domain.URL, error)
	GetOriginalURL(shortCode string) (string, error)
	GetAllURLS() ([]domain.URL, error)
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Handler
type Handler struct {
	log     *slog.Logger
	service URLService
	baseUrl string
}

// Handler constructor
func NewHandler(log *slog.Logger, service URLService, baseUrl string) *Handler {
	return &Handler{
		log:     log,
		service: service,
		baseUrl: baseUrl,
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	metrics.URLShortenRequests.Inc()
	timer := prometheus.NewTimer(metrics.RequestDuration)
	defer timer.ObserveDuration()

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	urls, err := h.service.ShortenURL(req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to shorten URL", http.StatusInternalServerError)
		return
	}

	path, err := url.JoinPath(h.baseUrl, urls.ShortCode)
	if err != nil {
		http.Error(w, "create short url path", http.StatusInternalServerError)
		return
	}

	res := URLResponse{
		ShortURL:    path,
		OriginalURL: urls.OriginalURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.log.Error("endcode struct", "error", err)
	}
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	metrics.URLRedirectRequests.Inc()
	timer := prometheus.NewTimer(metrics.RequestDuration)
	defer timer.ObserveDuration()

	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	originalURL, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		if errors.Is(err, domain.ErrURLNotFound) {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	metrics.URLAccessCount.WithLabelValues(shortCode).Inc()

	redirectURL := originalURL

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) GetURLs(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.GetAllURLS()
	if err != nil {
		http.Error(w, "Failed to get URLs", http.StatusInternalServerError)
		return
	}

	var urls []URLResponse

	for _, url := range list {

		urls = append(urls, URLResponse{
			ShortURL:    h.baseUrl + "/" + url.ShortCode,
			OriginalURL: url.OriginalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}
