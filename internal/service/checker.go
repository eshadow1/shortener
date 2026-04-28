package service

import (
	"context"

	"github.com/eshadow1/shortener/internal/loggers"

	_ "github.com/lib/pq"
)

type repoChecker interface {
	PingContext(ctx context.Context) error
}

type checkerService struct {
	repo repoChecker
}

func NewCheckerService(r repoChecker) *checkerService {
	return &checkerService{
		repo: r,
	}
}

func (cs *checkerService) CheckDB(ctx context.Context) bool {
	if cs.repo == nil {
		loggers.Log.Info("Not used database")
		return false
	}

	if err := cs.repo.PingContext(ctx); err != nil {
		loggers.Log.Info("Not connected to database")
		return false
	}

	loggers.Log.Info("Connected to database")
	return true
}
