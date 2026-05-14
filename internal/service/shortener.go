package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
)

type Repository interface {
	Save(ctx context.Context, values []model.URLInfo) error
	Get(ctx context.Context, key string) (model.UserURL, error)
	GetUserURLs(ctx context.Context) ([]model.UserURL, error)
	DeleteUserURLs(ctx context.Context, userID string, urls []string) error
	Close()
}

type shortenerService struct {
	repo      Repository
	wg        sync.WaitGroup
	ctx       context.Context
	cancelCtx context.CancelFunc
	input     chan model.DeleteInfo
	batchSize int
}

var (
	ErrorDeleteShortURL = errors.New("failed to delete short url")
	ErrorAddToDeleteURL = errors.New("failed to add query to delete")
)

func NewShortenerService(repo Repository, cfg configs.ServiceConfig) *shortenerService {
	ctx, cancel := context.WithCancel(context.Background())

	s := &shortenerService{
		repo:      repo,
		ctx:       ctx,
		input:     make(chan model.DeleteInfo, cfg.BufferSizeChan),
		cancelCtx: cancel,
		batchSize: cfg.BatchSize,
	}

	s.wg.Add(1)
	go s.batchWorker()

	return s
}

func (s *shortenerService) batchWorker() {
	defer s.wg.Done()

	batch := make(map[string][]string)

	for {
		select {
		case req, ok := <-s.input:
			if !ok {
				return
			}
			batch[req.UserID] = append(batch[req.UserID], req.URLs...)

			if len(batch[req.UserID]) >= s.batchSize {
				s.flushBatch(req.UserID, batch[req.UserID])
				delete(batch, req.UserID)
			}
		case <-s.ctx.Done():
			for userID, urls := range batch {
				if len(urls) > 0 {
					s.flushBatch(userID, urls)
				}
			}
			return
		}
	}
}

func (dq *shortenerService) flushBatch(userID string, shortURLs []string) {
	if err := dq.repo.DeleteUserURLs(dq.ctx, userID, shortURLs); err != nil {
		loggers.Log.Errorf("failed to delete short urls: %v", err)
	}
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

	if origin.IsDeleted {
		return model.OriginalInfo{}, ErrorDeleteShortURL
	}

	return model.OriginalInfo{OriginalURL: origin.OriginalURL}, nil
}

func (s *shortenerService) GetUserURLs(ctx context.Context) ([]model.UserURL, error) {
	return s.repo.GetUserURLs(ctx)
}

func (s *shortenerService) DeleteUserShortURLs(ctx context.Context, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	select {
	case s.input <- model.DeleteInfo{UserID: ctx.Value(model.UserIDContextKey).(string), URLs: urls}:
		return nil
	default:
		loggers.Log.Errorf("error delete: %v", ErrorAddToDeleteURL)
		return nil
	}
}

func (s *shortenerService) Close() {
	s.cancelCtx()
	close(s.input)
	s.wg.Wait()
}
