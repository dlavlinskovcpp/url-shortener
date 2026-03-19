package usecase

import (
	"context"
	"errors"
	"testing"

	"shortener/internal/domain"
	"shortener/internal/repository/memory"
)

type generatorMock struct {
	values []string
	err    error
	index  int
}

func (g *generatorMock) Generate(length int) (string, error) {
	if g.err != nil {
		return "", g.err
	}
	if g.index >= len(g.values) {
		return "", errors.New("no more generated values")
	}
	v := g.values[g.index]
	g.index++
	return v, nil
}

type storageMock struct {
	getByOriginalFn func(ctx context.Context, original string) (string, error)
	getOriginalFn   func(ctx context.Context, short string) (string, error)
	createFn        func(ctx context.Context, original, short string) error
}

func (m *storageMock) GetByOriginal(ctx context.Context, original string) (string, error) {
	return m.getByOriginalFn(ctx, original)
}

func (m *storageMock) GetOriginal(ctx context.Context, short string) (string, error) {
	return m.getOriginalFn(ctx, short)
}

func (m *storageMock) Create(ctx context.Context, original, short string) error {
	return m.createFn(ctx, original, short)
}

func TestURLUseCase_Shorten(t *testing.T) {
	storage := memory.NewMemoryRepo()
	gen := &generatorMock{values: []string{
		"Abc123_XyZ",
		"Qwerty_123",
		"ZXCVBN_987",
	}}
	uc := NewURLUseCase(storage, gen, 5)
	ctx := context.Background()

	t.Run("valid url returns short", func(t *testing.T) {
		orig := "https://golang.org"

		short, err := uc.Shorten(ctx, orig)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(short) != 10 {
			t.Fatalf("expected short length 10, got %d", len(short))
		}
	})

	t.Run("same url returns same short", func(t *testing.T) {
		orig := "https://alpha.dev/path"

		short1, err := uc.Shorten(ctx, orig)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		short2, err := uc.Shorten(ctx, orig)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if short1 != short2 {
			t.Fatalf("expected same short, got %q and %q", short1, short2)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		_, err := uc.Shorten(ctx, "invalid-url")
		if !errors.Is(err, domain.ErrInvalidURL) {
			t.Fatalf("expected ErrInvalidURL, got %v", err)
		}
	})

	t.Run("invalid scheme", func(t *testing.T) {
		_, err := uc.Shorten(ctx, "ftp://example.com/file")
		if !errors.Is(err, domain.ErrInvalidURL) {
			t.Fatalf("expected ErrInvalidURL, got %v", err)
		}
	})
}

func TestURLUseCase_GetOriginal(t *testing.T) {
	storage := memory.NewMemoryRepo()
	gen := &generatorMock{values: []string{"Abc123_XyZ"}}
	uc := NewURLUseCase(storage, gen, 5)
	ctx := context.Background()

	orig := "https://service.test/resource"

	short, err := uc.Shorten(ctx, orig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("existing short", func(t *testing.T) {
		got, err := uc.GetOriginal(ctx, short)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != orig {
			t.Fatalf("expected %q, got %q", orig, got)
		}
	})

	t.Run("not existing short", func(t *testing.T) {
		_, err := uc.GetOriginal(ctx, "Z9Y8X7W6V_")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("invalid short too short", func(t *testing.T) {
		_, err := uc.GetOriginal(ctx, "abc")
		if !errors.Is(err, domain.ErrInvalidShort) {
			t.Fatalf("expected ErrInvalidShort, got %v", err)
		}
	})

	t.Run("invalid short characters", func(t *testing.T) {
		_, err := uc.GetOriginal(ctx, "!!!!!!!!!1")
		if !errors.Is(err, domain.ErrInvalidShort) {
			t.Fatalf("expected ErrInvalidShort, got %v", err)
		}
	})
}

func TestURLUseCase_Shorten_RetryOnShortCollision(t *testing.T) {
	ctx := context.Background()
	gen := &generatorMock{
		values: []string{"COLLIDE_01", "UNIQUE___2"},
	}

	var createCalls int
	storage := &storageMock{
		getByOriginalFn: func(ctx context.Context, original string) (string, error) {
			return "", domain.ErrNotFound
		},
		createFn: func(ctx context.Context, original, short string) error {
			createCalls++
			if createCalls == 1 {
				return domain.ErrShortExists
			}
			return nil
		},
		getOriginalFn: func(ctx context.Context, short string) (string, error) {
			return "", domain.ErrNotFound
		},
	}

	uc := NewURLUseCase(storage, gen, 2)

	short, err := uc.Shorten(ctx, "https://alpha.dev/retry")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if short != "UNIQUE___2" {
		t.Fatalf("expected UNIQUE___2, got %q", short)
	}
}

func TestURLUseCase_Shorten_ReturnsExistingAfterOriginalConflict(t *testing.T) {
	ctx := context.Background()
	gen := &generatorMock{
		values: []string{"Abc123_XyZ"},
	}

	getByOriginalCalls := 0
	storage := &storageMock{
		getByOriginalFn: func(ctx context.Context, original string) (string, error) {
			getByOriginalCalls++
			if getByOriginalCalls == 1 {
				return "", domain.ErrNotFound
			}
			return "EXISTING_01", nil
		},
		createFn: func(ctx context.Context, original, short string) error {
			return domain.ErrOriginalExists
		},
		getOriginalFn: func(ctx context.Context, short string) (string, error) {
			return "", domain.ErrNotFound
		},
	}

	uc := NewURLUseCase(storage, gen, 3)

	short, err := uc.Shorten(ctx, "https://alpha.dev/race")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if short != "EXISTING_01" {
		t.Fatalf("expected EXISTING_01, got %q", short)
	}
}

func TestURLUseCase_Shorten_GeneratorError(t *testing.T) {
	ctx := context.Background()
	storage := &storageMock{
		getByOriginalFn: func(ctx context.Context, original string) (string, error) {
			return "", domain.ErrNotFound
		},
		createFn: func(ctx context.Context, original, short string) error {
			return nil
		},
		getOriginalFn: func(ctx context.Context, short string) (string, error) {
			return "", domain.ErrNotFound
		},
	}

	gen := &generatorMock{err: errors.New("rand failure")}
	uc := NewURLUseCase(storage, gen, 3)

	_, err := uc.Shorten(ctx, "https://alpha.dev/generator")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestURLUseCase_Shorten_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	gen := &generatorMock{
		values: []string{"AAAAAAAAAA", "BBBBBBBBBB", "CCCCCCCCCC"},
	}
	storage := &storageMock{
		getByOriginalFn: func(ctx context.Context, original string) (string, error) {
			return "", domain.ErrNotFound
		},
		createFn: func(ctx context.Context, original, short string) error {
			return domain.ErrShortExists
		},
		getOriginalFn: func(ctx context.Context, short string) (string, error) {
			return "", domain.ErrNotFound
		},
	}

	uc := NewURLUseCase(storage, gen, 3)

	_, err := uc.Shorten(ctx, "https://alpha.dev/retries")
	if !errors.Is(err, domain.ErrShortGeneration) {
		t.Fatalf("expected ErrShortGeneration, got %v", err)
	}
}
