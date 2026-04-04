package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vadim-Makhnev/url-shortener/internal/domain"
	"github.com/Vadim-Makhnev/url-shortener/internal/handler/mocks"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandler_ShortenURL_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	handler := NewHandler(newTestLogger(), nil, "")

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", bytes.NewBufferString(`{"url":`))
	rec := httptest.NewRecorder()

	handler.ShortenURL(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_ShortenURL_ServiceInvalidURL_ReturnsBadRequest(t *testing.T) {
	svc := mocks.NewURLService(t)

	handler := NewHandler(newTestLogger(), svc, "http://localhost:8080")

	svc.EXPECT().
		ShortenURL("bad-url").Return(nil, domain.ErrInvalidURL)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", bytes.NewBufferString(`{"url":"bad-url"}`))
	rec := httptest.NewRecorder()

	handler.ShortenURL(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_ShortenURL_Success_ReturnsCreatedWithJSON(t *testing.T) {
	svc := mocks.NewURLService(t)

	handler := NewHandler(newTestLogger(), svc, "http://localhost:8080")
	original := "https://example.com"

	svc.EXPECT().
		ShortenURL(original).
		Return(&domain.URL{
			ShortCode:   "abc123",
			OriginalURL: original,
		}, nil)

	reqBody := fmt.Sprintf(`{"url": "%s"}`, original)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	var urlResp struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	handler.ShortenURL(rec, req)
	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	require.Equal(t, http.StatusCreated, rec.Code)

	err := json.NewDecoder(rec.Body).Decode(&urlResp)
	require.NoError(t, err)
	require.Equal(t, original, urlResp.OriginalURL)
	require.Equal(t, "http://localhost:8080/abc123", urlResp.ShortURL)
}

func TestHandler_RedirectURL_Success_ReturnsFoundWithLocation(t *testing.T) {
	svc := mocks.NewURLService(t)

	handler := NewHandler(newTestLogger(), svc, "http://localhost:8080")

	svc.EXPECT().
		GetOriginalURL("abc123").Return("https://example.com", nil)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/abc123", nil)
	req.SetPathValue("shortCode", "abc123")
	rec := httptest.NewRecorder()

	handler.RedirectURL(rec, req)

	require.Equal(t, "https://example.com", rec.Header().Get("Location"))
	require.Equal(t, http.StatusFound, rec.Code)
}

func TestHandler_RedirectURL_NotFound_Returns404(t *testing.T) {
	svc := mocks.NewURLService(t)

	handler := NewHandler(newTestLogger(), svc, "http://localhost:8080")

	svc.EXPECT().
		GetOriginalURL("abc123").Return("", domain.ErrURLNotFound)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/abc123", nil)
	req.SetPathValue("shortCode", "abc123")
	rec := httptest.NewRecorder()

	handler.RedirectURL(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "URL not found")
}

func TestHandler_GetURLs_Success_ReturnsJSONList(t *testing.T) {
	svc := mocks.NewURLService(t)

	handler := NewHandler(newTestLogger(), svc, "http://localhost:8080")

	svc.EXPECT().
		GetAllURLS().Return([]domain.URL{
		{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
		},
		{
			ShortCode:   "ad7fgd",
			OriginalURL: "https://google.com",
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/urls", nil)
	rec := httptest.NewRecorder()

	handler.GetURLs(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var got []URLResponse
	err := json.NewDecoder(rec.Body).Decode(&got)
	require.NoError(t, err)

	want := []URLResponse{
		{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
		{ShortURL: "http://localhost:8080/ad7fgd", OriginalURL: "https://google.com"},
	}

	require.Equal(t, want, got)
}
