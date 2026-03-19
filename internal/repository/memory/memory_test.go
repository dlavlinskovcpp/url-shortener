package memory

import (
	"context"
	"errors"
	"sync"
	"testing"

	"shortener/internal/domain"
)

func TestMemoryRepo_RoundTrip(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	original := "https://alpha.dev/page"
	short := "WF2N6410ZQ"

	err := repo.Create(ctx, original, short)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	gotShort, err := repo.GetByOriginal(ctx, original)
	if err != nil {
		t.Fatalf("get by original: %v", err)
	}
	if gotShort != short {
		t.Fatalf("short mismatch: %q", gotShort)
	}

	gotOriginal, err := repo.GetOriginal(ctx, short)
	if err != nil {
		t.Fatalf("get original: %v", err)
	}
	if gotOriginal != original {
		t.Fatalf("original mismatch: %q", gotOriginal)
	}
}

func TestMemoryRepo_GetByOriginal_Miss(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	_, err := repo.GetByOriginal(ctx, "https://lost.local/page")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestMemoryRepo_GetOriginal_Miss(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	_, err := repo.GetOriginal(ctx, "N4xT9mQ2Z_")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestMemoryRepo_Create_OriginalExists(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	err := repo.Create(ctx, "https://alpha.dev", "Q7Lp2Vx8Ks")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = repo.Create(ctx, "https://alpha.dev", "v8R1cT6YpQ")
	if !errors.Is(err, domain.ErrOriginalExists) {
		t.Fatalf("want ErrOriginalExists, got %v", err)
	}
}

func TestMemoryRepo_Create_ShortExists(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	err := repo.Create(ctx, "https://alpha.dev/1", "K9m2Pq7Lx_")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = repo.Create(ctx, "https://alpha.dev/2", "K9m2Pq7Lx_")
	if !errors.Is(err, domain.ErrShortExists) {
		t.Fatalf("want ErrShortExists, got %v", err)
	}
}

func TestMemoryRepo_ConcurrentCreateSameOriginal(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	const workers = 100
	original := "https://alpha.dev/shared"
	short := "N4xT9mQ2Z_"

	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- repo.Create(ctx, original, short)
		}()
	}

	wg.Wait()
	close(errCh)

	var okCount int
	var existsCount int

	for err := range errCh {
		switch {
		case err == nil:
			okCount++
		case errors.Is(err, domain.ErrOriginalExists):
			existsCount++
		default:
			t.Fatalf("unexpected result: %v", err)
		}
	}

	if okCount != 1 {
		t.Fatalf("ok count: %d", okCount)
	}

	if existsCount != workers-1 {
		t.Fatalf("exists count: %d", existsCount)
	}

	gotOriginal, err := repo.GetOriginal(ctx, short)
	if err != nil {
		t.Fatalf("get original: %v", err)
	}
	if gotOriginal != original {
		t.Fatalf("original mismatch: %q", gotOriginal)
	}
}
