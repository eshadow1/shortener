package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"sync"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/eshadow1/shortener/internal/pool"
)

var sha256Pool = pool.New[hash.Hash](sha256.New)

// Repository описывает контракт хранилища данных для операций создания,
// получения и удаления сокращенных URL-адресов.
type Repository interface {
	Save(ctx context.Context, values []model.URLInfo) error
	Get(ctx context.Context, key string) (model.UserURL, error)
	GetUserURLs(ctx context.Context) ([]model.UserURL, error)
	DeleteUserURLs(ctx context.Context, userID string, urls []string) error
	Close()
}

type shortenerService struct {
	repo          Repository
	wg            sync.WaitGroup
	ctx           context.Context
	cancelCtx     context.CancelFunc
	input         chan model.DeleteInfo
	batchSize     int
	flushInterval time.Duration
}

var (
	// ErrorDeleteShortURL возвращается при попытке доступа к короткому URL,
	// который был помечен как удалённый.
	ErrorDeleteShortURL = errors.New("failed to delete short url")
	// ErrorAddToDeleteURL возвращается, когда не удается добавить запрос на удаление в очередь.
	ErrorAddToDeleteURL = errors.New("failed to add query to delete")
)

// NewShortenerService создает и возвращает новый сервис для работы с сокращением URL,
// инициализируя фоновый воркер для пакетного удаления.
func NewShortenerService(repo Repository, cfg configs.ServiceConfig) *shortenerService {
	ctx, cancel := context.WithCancel(context.Background())

	s := &shortenerService{
		repo:          repo,
		ctx:           ctx,
		input:         make(chan model.DeleteInfo, cfg.BufferSizeChan),
		cancelCtx:     cancel,
		batchSize:     cfg.BatchSize,
		flushInterval: cfg.FlushInterval,
	}

	s.wg.Add(1)
	go s.batchWorker()

	return s
}

// batchWorker выполняет фоновую обработку очереди удалений, накапливая запросы в батчи.
func (s *shortenerService) batchWorker() {
	defer s.wg.Done()

	batch := make(map[string][]string)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

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
		case <-ticker.C:
			for userID, urls := range batch {
				if len(urls) > 0 {
					s.flushBatch(userID, urls)
				}
			}
			batch = make(map[string][]string)
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

// flushBatch выполняет немедленную отправку накопленных UR
func (dq *shortenerService) flushBatch(userID string, shortURLs []string) {
	if err := dq.repo.DeleteUserURLs(dq.ctx, userID, shortURLs); err != nil {
		loggers.Log.Errorf("failed to delete short urls: %v", err)
	}
}

// hashToShort генерирует короткое представление URL
func (*shortenerService) hashToShort(input string) string {
	data := CalculateHashWithPool(input)
	return hex.EncodeToString(data[:])[:8]
}

// CreateShortURL создает короткие URL для переданного списка оригинальных UR
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

// GetOriginalURL извлекает оригинальный URL по его короткому варианту из репозитория.
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

// GetUserURLs возвращает список всех URL-пар, принадлежащих текущему пользователю, из репозитория.
func (s *shortenerService) GetUserURLs(ctx context.Context) ([]model.UserURL, error) {
	return s.repo.GetUserURLs(ctx)
}

// DeleteUserShortURLs добавляет запрос на массовое удаление коротких URL в асинхронную очередь обработки.
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

// Close выполняет корректное завершение работы сервиса
func (s *shortenerService) Close() {
	s.cancelCtx()
	close(s.input)
	s.wg.Wait()
}

// CalculateHashWithPool - аналог sha256.Sum256, но с использованием пула
func CalculateHashWithPool(input string) [32]byte {
	h := sha256Pool.Get()
	defer sha256Pool.Put(h)

	h.Write([]byte(input))

	var res [32]byte
	copy(res[:], h.Sum(nil))

	return res
}
