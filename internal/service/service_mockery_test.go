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

	postgres.AssertNotCalled(t, "GetURLByShortCode", mock.Anything)
}

func TestURLService_ShortenURL_InvalidURL_ReturnsDomainInvalidURL(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	got, err := svc.ShortenURL("abc123")
	require.Nil(t, got)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInvalidURL))
}

func TestURLService_ShortenURL_CreateURLError_ReturnsDomainInternal(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	postgres.EXPECT().
		CreateURL(mock.AnythingOfType("string"), "https://example.com").
		Return(nil, errors.New("db down"))

	got, err := svc.ShortenURL("https://example.com")
	require.Nil(t, got)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInternal))

	redis.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything)
}

func TestURLService_GetOriginalURL_RedisError_FallbackToPostgresSuccess(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	redis.EXPECT().
		Get(mock.Anything, "abc123").Return("", repository.ErrInternal)

	postgres.EXPECT().
		GetURLByShortCode("abc123").Return("https://example.com", nil)

	got, err := svc.GetOriginalURL("abc123")
	redis.AssertCalled(t, "Get", mock.Anything, mock.Anything)
	require.Equal(t, "https://example.com", got)
	require.NoError(t, err)
}

func TestURLService_GetAllURLS_Success(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	postgres.EXPECT().
		GetAllURLS().Return([]domain.URL{
		{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
		},
		{
			ShortCode:   "qwe123",
			OriginalURL: "https://fake.com",
		},
	}, nil)

	expected := []domain.URL{
		{
			ShortCode:   "abc123",
			OriginalURL: "https://example.com",
		},
		{
			ShortCode:   "qwe123",
			OriginalURL: "https://fake.com",
		},
	}

	urls, err := svc.GetAllURLS()
	require.NoError(t, err)
	require.Equal(t, expected, urls)
}

func TestURLService_GetAllURLS_PostgresError_ReturnsDomainInternal(t *testing.T) {
	postgres := mocks.NewRepositoryPostgres(t)
	redis := mocks.NewRepositoryRedis(t)

	svc := NewService(newTestLogger(), postgres, redis)

	postgres.EXPECT().
		GetAllURLS().Return(nil, errors.New("db down"))

	urls, err := svc.GetAllURLS()
	require.Nil(t, urls)
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInternal))
}
