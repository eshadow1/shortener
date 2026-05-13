package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultDriver               = "postgres"
	defaultMaxIdleConnections   = 5
	defaultMaxOpenConnections   = 20
	defaultConnMaxLifetime      = 1 * time.Minute
	codePostgresDuplicateInsert = "23505"
	errorNoRows                 = "no rows in result set"
)

type postgreSQLRepository struct {
	db *sql.DB
}

func NewPostgreSQLRepository(cfg configs.StorageConfig) (*postgreSQLRepository, error) {
	db, errOpen := sql.Open("pgx", cfg.PathDB)
	if errOpen != nil {
		return nil, fmt.Errorf("error create PostgreSQL DB: %w", errOpen)
	}

	db.SetMaxOpenConns(defaultMaxOpenConnections)
	db.SetMaxIdleConns(defaultMaxIdleConnections)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)

	if errMigrate := runMigrationsWithDB(db, "file://"+cfg.PathMigrations); errMigrate != nil {
		return nil, fmt.Errorf("error migrate: %w", errMigrate)
	}
	loggers.Log.Info("Migrate successful")

	return &postgreSQLRepository{
		db: db,
	}, nil
}

func (repo *postgreSQLRepository) PingContext(ctx context.Context) error {
	return repo.db.PingContext(ctx)
}

func (repo *postgreSQLRepository) Save(ctx context.Context, values []model.URLInfo) error {
	const query = `
        INSERT INTO shorten (shorten_url, original_url, user_id)
        VALUES ($1, $2, $3) 
        ON CONFLICT (original_url) DO NOTHING
		RETURNING id;
    `

	tx, errBegin := repo.db.BeginTx(ctx, nil)
	if errBegin != nil {
		return errBegin
	}
	defer func() {
		if errRollBack := tx.Rollback(); errRollBack != nil && !errors.Is(errRollBack, sql.ErrTxDone) {
			loggers.Log.Errorf("failed insert transaction: %v", errRollBack)
		}
	}()

	userID := ctx.Value(model.UserIDContextKey).(string)

	for _, value := range values {
		var id int64
		errTransaction := tx.QueryRowContext(ctx, query, value.ShortURL, value.OriginalURL, userID).Scan(&id)
		if errTransaction != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](errTransaction); ok && pgErr.Code == codePostgresDuplicateInsert {
				return &model.CustomPostgresError{
					Message: "value already exists: ",
					Err:     errTransaction,
				}
			}

			if errors.Is(errTransaction, sql.ErrNoRows) || strings.Contains(errTransaction.Error(), errorNoRows) {
				return &model.CustomPostgresError{Message: "value already exists: ", Err: errTransaction}
			}

			return fmt.Errorf("failed to insert URL: %w", errTransaction)
		}
	}

	return tx.Commit()
}

func (repo *postgreSQLRepository) Get(ctx context.Context, key string) (string, error) {
	const query = `
        SELECT original_url 
        FROM shorten 
        WHERE shorten_url = $1 and user_id = $2;
    `

	userID := ctx.Value(model.UserIDContextKey).(string)

	var shortenURL string
	err := repo.db.QueryRowContext(ctx, query, key, userID).Scan(&shortenURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("failed to query shorten_url: %w", err)
	}

	return shortenURL, nil
}

func (repo *postgreSQLRepository) GetUserURLs(ctx context.Context) ([]model.UserURL, error) {
	const query = `
        SELECT original_url, shorten_url 
        FROM shorten 
        WHERE user_id = $1;
    `

	userID := ctx.Value(model.UserIDContextKey).(string)

	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]model.UserURL, 0), sql.ErrNoRows
		}
		return make([]model.UserURL, 0), fmt.Errorf("failed to query shorten_url: %w", err)
	}

	urls := make([]model.UserURL, 0)

	for rows.Next() {
		var url model.UserURL
		if errScan := rows.Scan(&url.OriginalURL, &url.ShortURL); errScan != nil {
			return make([]model.UserURL, 0), fmt.Errorf("failed to query shorten_url: %w", err)
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (repo *postgreSQLRepository) Close() {
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
