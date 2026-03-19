package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"shortener/internal/domain"
)

const shortCodeLength = 10

type URLStorage interface {
	GetByOriginal(ctx context.Context, original string) (string, error)
	GetOriginal(ctx context.Context, short string) (string, error)
	Create(ctx context.Context, original, short string) error
}

type ShortGenerator interface {
	Generate(length int) (string, error)
}

type URLUseCase struct {
	storage    URLStorage
	generator  ShortGenerator
	maxRetries int
}

func NewURLUseCase(storage URLStorage, generator ShortGenerator, maxRetries int) *URLUseCase {
	if maxRetries < 1 {
		maxRetries = 1
	}

	return &URLUseCase{
		storage:    storage,
		generator:  generator,
		maxRetries: maxRetries,
	}
}

func (u *URLUseCase) Shorten(ctx context.Context, original string) (string, error) {
	normalized, err := validateAndNormalizeURL(original)
	if err != nil {
		return "", err
	}

	existingShort, err := u.storage.GetByOriginal(ctx, normalized)
	if err == nil {
		return existingShort, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return "", fmt.Errorf("lookup original url: %w", err)
	}

	for i := 0; i < u.maxRetries; i++ {
		short, err := u.generator.Generate(shortCodeLength)
		if err != nil {
			return "", fmt.Errorf("generate short: %w", err)
		}

		err = u.storage.Create(ctx, normalized, short)
		if err == nil {
			return short, nil
		}

		if errors.Is(err, domain.ErrShortExists) {
			continue
		}

		if errors.Is(err, domain.ErrOriginalExists) {
			existingShort, getErr := u.storage.GetByOriginal(ctx, normalized)
			if getErr != nil {
				return "", fmt.Errorf("fetch existing short after original conflict: %w", getErr)
			}
			return existingShort, nil
		}

		return "", fmt.Errorf("store short url: %w", err)
	}

	return "", domain.ErrShortGeneration
}

func (u *URLUseCase) GetOriginal(ctx context.Context, short string) (string, error) {
	short = strings.TrimSpace(short)
	if err := validateShortCode(short); err != nil {
		return "", err
	}

	original, err := u.storage.GetOriginal(ctx, short)
	if err != nil {
		return "", fmt.Errorf("lookup original by short: %w", err)
	}

	return original, nil
}

func validateAndNormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", domain.ErrInvalidURL
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", domain.ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", domain.ErrInvalidURL
	}

	if parsed.Host == "" {
		return "", domain.ErrInvalidURL
	}

	parsed.Host = strings.ToLower(parsed.Host)

	return parsed.String(), nil
}

func validateShortCode(short string) error {
	if len(short) != shortCodeLength {
		return domain.ErrInvalidShort
	}

	for _, r := range short {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' {
			continue
		}
		return domain.ErrInvalidShort
	}

	return nil
}
