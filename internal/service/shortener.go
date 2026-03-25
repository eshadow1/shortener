package service

import (
	"crypto/md5"
	"encoding/hex"
)

type repository interface {
	Save(key, value string) error
	Get(key string) (string, error)
}
type shortenerService struct {
	repo repository
}

func NewShortenerService(repo repository) *shortenerService {
	return &shortenerService{repo: repo}
}

func (*shortenerService) hashToShort(input string) string {
	hash := md5.Sum([]byte(input))

	return hex.EncodeToString(hash[:])[:8]
}

func (s *shortenerService) CreateShortUrl(original string) (string, error) {
	short := s.hashToShort(original)

	errSave := s.repo.Save(short, original)
	if errSave != nil {
		return short, errSave
	}

	return short, nil
}

func (s *shortenerService) GetOriginalURL(short string) (string, error) {
	return s.repo.Get(short)
}
