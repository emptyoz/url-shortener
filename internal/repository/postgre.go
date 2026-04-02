package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func NewPostgreSQL(log *slog.Logger, address string) (*DB, error) {
	conn, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: conn,
	}, nil
}

func (db *DB) CreateURL(shortCode, originalURL string) (*domain.URL, error) {
	const op = "repository.CreateURL"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO urls (short_code, original_url) 
			VALUES ($1, $2);`

	res, err := db.conn.ExecContext(ctx, query, shortCode, originalURL)
	if err != nil {
		db.log.Error("CreateURL", "original_url", originalURL, "short_code", shortCode, "error", err)
		return nil, fmt.Errorf("%s: create short url %w", op, errors.Join(ErrInternal, err))
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, errors.Join(ErrInternal, err))
	}

	if cnt == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrInternal)
	}

	return &domain.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}, nil
}

func (db *DB) GetURLByShortCode(shortCode string) (string, error) {
	const op = "repository.GetURLByShortCode"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var originalURL string
	query := `SELECT original_url FROM urls 
			WHERE short_code = $1;`

	err := db.conn.QueryRowContext(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: get original url: %w", op, errors.Join(ErrNotFound, err))
		}
		db.log.Error("GetURLByshortCode", "short_code", shortCode, "error", err)
		return "", fmt.Errorf("%s: %w", op, errors.Join(ErrInternal, err))
	}

	return originalURL, nil
}

func (db *DB) GetAllURLS() ([]domain.URL, error) {
	const op = "repository.GetAllURLs"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT short_code, original_url FROM urls ORDER BY
			created_at DESC;`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		db.log.Error("GetAllURLS", "error", err)
		return nil, fmt.Errorf("%s: get all urls %w", op, errors.Join(ErrInternal, err))
	}
	defer rows.Close()

	var urls []domain.URL
	for rows.Next() {
		var url domain.URL
		if err := rows.Scan(&url.ShortCode, &url.OriginalURL); err != nil {
			db.log.Error("GetAllURLS scan", "error", err)
			return nil, fmt.Errorf("%s: %w", op, errors.Join(ErrInternal, err))
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		db.log.Error("GetAllURLS rows", "error", err)
		return nil, fmt.Errorf("%s: %w", op, errors.Join(ErrInternal, err))
	}

	return urls, nil
}
