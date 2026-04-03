package service

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/Vadim-Makhnev/url-shortener/internal/domain"
	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
	"github.com/Vadim-Makhnev/url-shortener/internal/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestURLService_ShortenURL_Success(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)
	original := "https://example.com"

	postgres.EXPECT().
		CreateURL(mock.AnythingOfType("string"), original).
		RunAndReturn(func(shortCode, originalURL string) (*domain.URL, error) {
			return &domain.URL{
				ShortCode:   shortCode,
				OriginalURL: originalURL,
			}, nil
		})

	redis.EXPECT().
		Set(mock.Anything, mock.AnythingOfType("string"), original).
		Return(nil)

	got, err := svc.ShortenURL(original)
	require.NoError(t, err)
	require.Equal(t, original, got.OriginalURL)
	require.Len(t, got.ShortCode, 6)
}

func TestURLService_ShortenURL_RedisSetError(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)
	original := "https://example.com"

	postgres.EXPECT().
		CreateURL(mock.AnythingOfType("string"), original).
		RunAndReturn(func(shortCode, originalURL string) (*domain.URL, error) {
			return &domain.URL{
				ShortCode:   shortCode,
				OriginalURL: originalURL,
			}, nil
		})

	redis.EXPECT().
		Set(mock.Anything, mock.AnythingOfType("string"), original).
		Return(errors.New("redis unavailable"))

	got, err := svc.ShortenURL(original)
	require.Nil(t, got)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInternal))
}

func TestURLService_GetOriginalURL_PostgresNotFound_ReturnsDomainNotFound(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	postgres.EXPECT().
		GetURLByShortCode(mock.AnythingOfType("string")).
		RunAndReturn(func(shortCode string) (string, error) {
			return "", repository.ErrNotFound
		})

	redis.EXPECT().
		Get(mock.Anything, mock.AnythingOfType("string")).
		Return("", repository.ErrNotFound)

	got, err := svc.GetOriginalURL("abc123")
	require.Empty(t, got)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrURLNotFound))
}

func TestURLService_GetOriginalURL_PostgresInternal_ReturnsDomainInternal(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	postgres.EXPECT().
		GetURLByShortCode(mock.AnythingOfType("string")).
		RunAndReturn(func(shortCode string) (string, error) {
			return "", repository.ErrInternal
		})

	redis.EXPECT().
		Get(mock.Anything, mock.AnythingOfType("string")).
		Return("", repository.ErrNotFound)

	got, err := svc.GetOriginalURL("abc123")
	require.Empty(t, got)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInternal))
}

func TestURLService_GetOriginalURL_RedisHit_SkipsPostgres(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	redis.EXPECT().
		Get(mock.Anything, mock.AnythingOfType("string")).
		Return("https://example.com", nil)

	got, err := svc.GetOriginalURL("abc123")
	require.NotEmpty(t, got)
	require.NoError(t, err)
	require.Equal(t, "https://example.com", got)
}
