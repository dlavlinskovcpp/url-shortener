package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"shortener/internal/domain"

	"github.com/lib/pq"
)

const (
	constraintOriginalURLUnique = "urls_original_url_key"
	constraintShortUnique       = "urls_short_key"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) GetByOriginal(ctx context.Context, original string) (string, error) {
	const query = `
		SELECT short
		FROM urls
		WHERE original_url = $1
	`

	var short string
	err := r.db.QueryRowContext(ctx, query, original).Scan(&short)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("get short by original: %w", err)
	}

	return short, nil
}

func (r *PostgresRepo) Create(ctx context.Context, original, short string) error {
	const query = `
		INSERT INTO urls (original_url, short)
		VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(ctx, query, original, short)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Constraint {
			case constraintOriginalURLUnique:
				return domain.ErrOriginalExists
			case constraintShortUnique:
				return domain.ErrShortExists
			}
		}
		return fmt.Errorf("create url mapping: %w", err)
	}

	return nil
}

func (r *PostgresRepo) GetOriginal(ctx context.Context, short string) (string, error) {
	const query = `
		SELECT original_url
		FROM urls
		WHERE short = $1
	`

	var original string
	err := r.db.QueryRowContext(ctx, query, short).Scan(&original)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("get original by short: %w", err)
	}

	return original, nil
}
