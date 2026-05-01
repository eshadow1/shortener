package service

import (
	"context"
	"fmt"

	"github.com/eshadow1/shortener/internal/loggers"
)

type RepoChecker interface {
	PingContext(ctx context.Context) error
}

type checkerService struct {
	repo RepoChecker
}

func NewCheckerService(r RepoChecker) *checkerService {
	return &checkerService{
		repo: r,
	}
}

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
