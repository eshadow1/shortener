package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/eshadow1/shortener/internal/model"
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

func (s *shortenerService) CreateShortUrl(ctx context.Context, original model.OriginalInfo) (model.ShortenInfo, error) {
	short := s.hashToShort(original.OriginalURL)

	errSave := s.repo.Save(ctx, short, original.OriginalURL)
	if errSave != nil {
		return model.ShortenInfo{
			ShortURL: short,
		}, errSave
	}

	return model.ShortenInfo{
		ShortURL: short,
	}, nil
}

func (s *shortenerService) GetOriginalURL(ctx context.Context, short model.ShortenInfo) (model.OriginalInfo, error) {
	origin, errGet := s.repo.Get(ctx, short.ShortURL)
	if errGet != nil {
		return model.OriginalInfo{}, errGet
	}

	return model.OriginalInfo{OriginalURL: origin}, nil
}
