package memory

import (
	"context"
	"shortener/internal/domain"
	"sync"
)

type MemoryRepo struct {
	mu          sync.RWMutex
	origToShort map[string]string
	shortToOrig map[string]string
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		origToShort: make(map[string]string),
		shortToOrig: make(map[string]string),
	}
}

func (r *MemoryRepo) GetByOriginal(_ context.Context, original string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	short, ok := r.origToShort[original]
	if !ok {
		return "", domain.ErrNotFound
	}

	return short, nil
}

func (r *MemoryRepo) Create(_ context.Context, original, short string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.origToShort[original]; ok {
		return domain.ErrOriginalExists
	}

	if _, ok := r.shortToOrig[short]; ok {
		return domain.ErrShortExists
	}

	r.origToShort[original] = short
	r.shortToOrig[short] = original

	return nil
}

func (r *MemoryRepo) GetOriginal(_ context.Context, short string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	original, ok := r.shortToOrig[short]
	if !ok {
		return "", domain.ErrNotFound
	}

	return original, nil
}
