package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

type repository interface {
	Save(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
}
type shortenerService struct {
	repo repository
}

func NewShortenerService(repo repository) *shortenerService {
	return &shortenerService{repo: repo}
}

func (*shortenerService) hashToShort(input string) string {
	hash := sha256.Sum256([]byte(input))

	return hex.EncodeToString(hash[:])[:8]
}

func (s *shortenerService) CreateShortUrl(ctx context.Context, original string) (string, error) {
	short := s.hashToShort(original)

	errSave := s.repo.Save(ctx, short, original)
	if errSave != nil {
		return short, errSave
	}

	return short, nil
}

func (s *shortenerService) GetOriginalURL(ctx context.Context, short string) (string, error) {
	return s.repo.Get(ctx, short)
}
