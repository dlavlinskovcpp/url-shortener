package domain

import "errors"

var (
	ErrNotFound        = errors.New("url not found")
	ErrInvalidURL      = errors.New("invalid original url")
	ErrInvalidShort    = errors.New("invalid short url")
	ErrOriginalExists  = errors.New("original url already exists")
	ErrShortExists     = errors.New("short url already exists")
	ErrShortGeneration = errors.New("failed to generate unique short url")
)
