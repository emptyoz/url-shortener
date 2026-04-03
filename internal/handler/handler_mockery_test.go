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

	require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
}

func TestHandler_ShortenURL_ServiceInvalidURL_ReturnsBadRequest(t *testing.T) {
	svc := mocks.NewURLService(t)

	handler := NewHandler(newTestLogger(), svc, "http://localhost:8080")

	svc.EXPECT().
		ShortenURL("bad-url").Return(nil, domain.ErrInvalidURL)

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", bytes.NewBufferString(`{"url":"bad-url"}`))
	rec := httptest.NewRecorder()

	handler.ShortenURL(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
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
	require.Equal(t, http.StatusCreated, rec.Result().StatusCode)

	err := json.NewDecoder(rec.Body).Decode(&urlResp)
	require.NoError(t, err)
	require.Equal(t, original, urlResp.OriginalURL)
	require.Equal(t, "http://localhost:8080/abc123", urlResp.ShortURL)
}
