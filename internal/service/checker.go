package service

import (
	"context"
	"fmt"

	"github.com/eshadow1/shortener/internal/loggers"
)

// RepoChecker описывает контракт для проверки доступности и работоспособности хранилища данных.
type RepoChecker interface {
	PingContext(ctx context.Context) error
}

type checkerService struct {
	repo RepoChecker
}

// NewCheckerService создает и возвращает новый сервис для проверки состояния подключения к базе данных.
func NewCheckerService(r RepoChecker) *checkerService {
	return &checkerService{
		repo: r,
	}
}

// ConnectDB выполняет проверку подключения к базе данных путем вызова метода PingContext у репозитория.
func (cs *checkerService) ConnectDB(ctx context.Context) error {
	if cs.repo == nil {
		return fmt.Errorf("not used database")
	}

	if err := cs.repo.PingContext(ctx); err != nil {
		return fmt.Errorf("not connected to database")
	}

	loggers.Log.Info("Connected to database")
	return nil
}
