package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	_ "github.com/lib/pq"
)

const (
	defaultMaxIdleConnections = 5
	defaultMaxOpenConnections = 20
	defaultConnMaxLifetime    = 1 * time.Minute
)

type PostgreSQLRepository struct {
	db *sql.DB
}

func NewPostgreSQLRepository(cfg configs.StorageConfig) (*PostgreSQLRepository, error) {
	db, errOpen := sql.Open("postgres", cfg.PathDB)
	if errOpen != nil {
		loggers.Log.Errorf("Ошибка в создании PostgreSQL DB: %v", errOpen)
		return nil, errOpen
	}

	db.SetMaxOpenConns(defaultMaxOpenConnections)
	db.SetMaxIdleConns(defaultMaxIdleConnections)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)

	return &PostgreSQLRepository{
		db: db,
	}, nil
}

func (repo *PostgreSQLRepository) PingContext(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

func (repo *PostgreSQLRepository) Save(_ context.Context, key, value string) error {
	return nil
}
func (repo *PostgreSQLRepository) Get(_ context.Context, key string) (string, error) {
	return "", nil
}
func (repo *PostgreSQLRepository) Close() {
	defer repo.db.Close()
}
