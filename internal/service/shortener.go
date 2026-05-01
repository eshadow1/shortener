package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/eshadow1/shortener/internal/model"
)

type Repository interface {
	Save(ctx context.Context, values []model.URLInfo) error
	Get(ctx context.Context, key string) (string, error)
	Close()
}
type shortenerService struct {
	repo Repository
}

func NewShortenerService(repo Repository) *shortenerService {
	return &shortenerService{repo: repo}
}

func (*shortenerService) hashToShort(input string) string {
	hash := sha256.Sum256([]byte(input))

	return hex.EncodeToString(hash[:])[:8]
}

func (s *shortenerService) CreateShortURL(ctx context.Context, originals []model.OriginalInfo) ([]model.ShortenInfo, error) {
	shortens := make([]model.ShortenInfo, 0, len(originals))
	urlsInfo := make([]model.URLInfo, 0, len(originals))
	for _, original := range originals {
		short := s.hashToShort(original.OriginalURL)

		shortens = append(shortens, model.ShortenInfo{
			ShortURL:      short,
			CorrelationID: original.CorrelationID,
		})

		urlsInfo = append(urlsInfo, model.URLInfo{
			OriginalURL: original.OriginalURL,
			ShortURL:    short,
		})
	}

	if len(urlsInfo) != 0 {
		errSave := s.repo.Save(ctx, urlsInfo)
		if errSave != nil {
			return shortens, errSave
		}
	}

	return shortens, nil
}

func (s *shortenerService) GetOriginalURL(ctx context.Context, short model.ShortenInfo) (model.OriginalInfo, error) {
	origin, errGet := s.repo.Get(ctx, short.ShortURL)
	if errGet != nil {
		return model.OriginalInfo{}, errGet
	}

	return model.OriginalInfo{OriginalURL: origin}, nil
}
