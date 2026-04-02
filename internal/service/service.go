package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	neturl "net/url"
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/domain"
	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
)

var (
	defaultTimeout = 5 * time.Second
)

type RepositoryPostgres interface {
	CreateURL(shortCode, originalURL string) (*domain.URL, error)
	GetURLByShortCode(shortCode string) (string, error)
	GetAllURLS() ([]domain.URL, error)
}

type RepositoryRedis interface {
	Set(ctx context.Context, shortCode, originalURL string) error
	Get(ctx context.Context, shortCode string) (string, error)
}

type URLService struct {
	postgres RepositoryPostgres
	logger   *slog.Logger
	redis    RepositoryRedis
}

func NewService(log *slog.Logger, repo RepositoryPostgres, redis RepositoryRedis) *URLService {
	return &URLService{
		logger:   log,
		postgres: repo,
		redis:    redis,
	}
}

func (s *URLService) ShortenURL(originalURL string) (*domain.URL, error) {
	const op = "service.ShortenURL"

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	parsed, err := neturl.ParseRequestURI(originalURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInvalidURL)
	}

	shortCode := generateShortCode()

	url, err := s.postgres.CreateURL(shortCode, originalURL)
	if err != nil {
		s.logger.Error("ShortenURL:", "error", err)
		return nil, fmt.Errorf("%s: create url: %w", op, domain.ErrInternal)
	}

	err = s.redis.Set(ctx, shortCode, originalURL)
	if err != nil {
		s.logger.Warn("ShortenURL cache set", "short_code", shortCode, "error", err)
		return nil, fmt.Errorf("%s: cache set: %w", op, domain.ErrInternal)
	}

	domainURL := &domain.URL{
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
	}

	return domainURL, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	const op = "service.GetOriginalURL"

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	val, err := s.redis.Get(ctx, shortCode)
	if err == nil {
		return val, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		s.logger.Warn("GetOriginalURL cache get", "short_code", shortCode, "error", err)
	}

	url, err := s.postgres.GetURLByShortCode(shortCode)
	if err != nil {
		s.logger.Error("GetOriginalURL:", "error", err)
		if errors.Is(err, repository.ErrNotFound) {
			return "", fmt.Errorf("%s: %w", op, domain.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, domain.ErrInternal)
	}

	return url, nil
}

func (s *URLService) GetAllURLS() ([]domain.URL, error) {
	const op = "service.GetAllURLs"

	urls, err := s.postgres.GetAllURLS()
	if err != nil {
		s.logger.Error("GetAllURLS:", "error", err)
		return nil, fmt.Errorf("%s: %w", op, domain.ErrInternal)
	}

	var res []domain.URL

	for _, url := range urls {
		res = append(res, domain.URL{
			ShortCode:   url.ShortCode,
			OriginalURL: url.OriginalURL,
		})
	}

	return res, nil
}

func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
