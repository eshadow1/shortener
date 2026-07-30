package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPermDir = 0644

func TestPostgreSQLRepository_NewPostgreSQLRepository(t *testing.T) {
	errLog := loggers.CreateLogger("error")
	require.NoError(t, errLog)

	dsn := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	tmpDir := t.TempDir()
	migrationDir := filepath.Join(tmpDir, "migrations")
	errCreate := os.Mkdir(migrationDir, testPermDir)
	require.NoError(t, errCreate)

	tests := []struct {
		name        string
		cfg         configs.StorageConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid DSN without migrations",
			cfg: configs.StorageConfig{
				PathDB:         dsn,
				PathMigrations: "",
			},
			wantErr: false,
		},
		{
			name: "valid DSN but migrations path does not exist",
			cfg: configs.StorageConfig{
				PathDB:         dsn,
				PathMigrations: "/nonexistent/path",
			},
			wantErr:     true,
			errContains: "error migrate",
		},
		{
			name: "valid DSN with empty migrations directory",
			cfg: configs.StorageConfig{
				PathDB:         dsn,
				PathMigrations: migrationDir,
			},
			wantErr:     true,
			errContains: "error migrate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, errNew := NewPostgreSQLRepository(tt.cfg)
			if tt.wantErr {
				require.Error(t, errNew)
				if tt.errContains != "" {
					assert.Contains(t, errNew.Error(), tt.errContains)
				}
				assert.Nil(t, repo)
			} else {
				require.NoError(t, errNew)
				assert.NotNil(t, repo)

				if repo.db != nil {
					repo.db.Close()
				}
				if repo.pool != nil {
					repo.pool.Close()
				}
			}
		})
	}
}
