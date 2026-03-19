package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"

	"shortener/internal/domain"
)

func TestPostgresRepo_GetByOriginal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepo(db)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"short"}).AddRow("WF2N6410ZQ")

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT short
		FROM urls
		WHERE original_url = $1
	`)).
			WithArgs("https://alpha.dev").
			WillReturnRows(rows)

		short, err := repo.GetByOriginal(ctx, "https://alpha.dev")
		if err != nil {
			t.Fatalf("get by original: %v", err)
		}
		if short != "WF2N6410ZQ" {
			t.Fatalf("short mismatch: %q", short)
		}
	})

	t.Run("miss", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT short
		FROM urls
		WHERE original_url = $1
	`)).
			WithArgs("https://missing.example").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetByOriginal(ctx, "https://missing.example")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPostgresRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepo(db)
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO urls (original_url, short)
		VALUES ($1, $2)
	`)).
			WithArgs("https://alpha.dev", "WF2N6410ZQ").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, "https://alpha.dev", "WF2N6410ZQ")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
	})

	t.Run("original exists", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO urls (original_url, short)
		VALUES ($1, $2)
	`)).
			WithArgs("https://duplicate.example", "K9m2Pq7Lx_").
			WillReturnError(&pq.Error{
				Code:       "23505",
				Constraint: constraintOriginalURLUnique,
			})

		err := repo.Create(ctx, "https://duplicate.example", "K9m2Pq7Lx_")
		if !errors.Is(err, domain.ErrOriginalExists) {
			t.Fatalf("want ErrOriginalExists, got %v", err)
		}
	})

	t.Run("short exists", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO urls (original_url, short)
		VALUES ($1, $2)
	`)).
			WithArgs("https://another.example", "Q7Lp2Vx8Ks").
			WillReturnError(&pq.Error{
				Code:       "23505",
				Constraint: constraintShortUnique,
			})

		err := repo.Create(ctx, "https://another.example", "Q7Lp2Vx8Ks")
		if !errors.Is(err, domain.ErrShortExists) {
			t.Fatalf("want ErrShortExists, got %v", err)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPostgresRepo_GetOriginal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock init: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepo(db)
	ctx := context.Background()

	t.Run("hit", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"original_url"}).AddRow("https://docs.local/manual")

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT original_url
		FROM urls
		WHERE short = $1
	`)).
			WithArgs("K9m2Pq7Lx_").
			WillReturnRows(rows)

		original, err := repo.GetOriginal(ctx, "K9m2Pq7Lx_")
		if err != nil {
			t.Fatalf("get original: %v", err)
		}
		if original != "https://docs.local/manual" {
			t.Fatalf("original mismatch: %q", original)
		}
	})

	t.Run("miss", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT original_url
		FROM urls
		WHERE short = $1
	`)).
			WithArgs("v8R1cT6YpQ").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetOriginal(ctx, "v8R1cT6YpQ")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
