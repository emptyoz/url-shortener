package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const address = "http://localhost:8080"

var client = http.Client{
	Timeout: 10 * time.Minute,
}

type Reply struct {
	Status string `json:"status"`
}

func TestHealth(t *testing.T) {
	resp, err := client.Get(address + "/health")
	require.NoError(t, err, "cannot health check")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var rep Reply
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rep))
	require.Equal(t, "OK", rep.Status)
}

func TestCreate_ShortenURL(t *testing.T) {
	reqBody := struct {
		URL string `json:"url"`
	}{
		URL: "https://google.com",
	}

	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, address+"/api/shorten", bytes.NewBuffer(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestGetURLS_GetURLs(t *testing.T) {
	created := []string{
		fmt.Sprintf("https://example.com/a/%d", time.Now().UnixNano()),
		fmt.Sprintf("https://example.com/b/%d", time.Now().UnixNano()+1),
	}

	for _, u := range created {
		createShortURL(t, u)
	}

	resp, err := client.Get(address + "/api/urls")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var got []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	require.GreaterOrEqual(t, len(got), len(created))

	gotOriginals := make(map[string]struct{}, len(got))
	for _, item := range got {
		require.NotEmpty(t, item.ShortURL)
		gotOriginals[item.OriginalURL] = struct{}{}
	}

	for _, want := range created {
		_, ok := gotOriginals[want]
		require.True(t, ok)
	}
}

func createShortURL(t *testing.T, original string) {
	t.Helper()

	reqBody := struct {
		URL string `json:"url"`
	}{
		URL: original,
	}

	b, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, address+"/api/shorten", bytes.NewBuffer(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}
