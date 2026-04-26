package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
)

const (
	defaultDriver             = "postgres"
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

	if errMigrate := runMigrationsWithDB(db, "file://"+cfg.PathMigrations); errMigrate != nil {
		loggers.Log.Errorf("Ошибка миграции: %v", errMigrate)
		return nil, errMigrate
	}
	loggers.Log.Info("Миграция выполнена")

	return &PostgreSQLRepository{
		db: db,
	}, nil
}

func (repo *PostgreSQLRepository) PingContext(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

func (repo *PostgreSQLRepository) Save(ctx context.Context, key, value string) error {
	const query = `
        INSERT INTO shorten (shorten_url, original_url)
        VALUES ($1, $2)
        RETURNING id
    `

	var id int64
	err := repo.db.QueryRowContext(ctx, query, key, value).Scan(&id)
	if err != nil {
		if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil
		}
		return fmt.Errorf("failed to insert URL: %w", err)
	}

	return nil
}

func (repo *PostgreSQLRepository) Get(ctx context.Context, key string) (string, error) {
	const query = `
        SELECT original_url 
        FROM shorten 
        WHERE shorten_url = $1
    `

	var shortenURL string
	err := repo.db.QueryRowContext(ctx, query, key).Scan(&shortenURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("failed to query shorten_url: %w", err)
	}

	return shortenURL, nil
}

func (repo *PostgreSQLRepository) Close() {
	repo.db.Close()
}

func runMigrationsWithDB(db *sql.DB, migrationsPath string) error {
	driver, errInstance := postgres.WithInstance(db, &postgres.Config{})
	if errInstance != nil {
		return fmt.Errorf("failed to create postgres driver: %w", errInstance)
	}

	m, errDBInstance := migrate.NewWithDatabaseInstance(
		migrationsPath,
		defaultDriver,
		driver,
	)
	if errDBInstance != nil {
		return fmt.Errorf("failed to init migrate: %w", errDBInstance)
	}

	if errUpMigrate := m.Up(); errUpMigrate != nil && !errors.Is(errUpMigrate, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", errUpMigrate)
	}

	return nil
}
