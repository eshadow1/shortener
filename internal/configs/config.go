// Package configs предоставляет функциональность для инициализации и управления
// конфигурацией приложения. Он поддерживает загрузку параметров через флаги
// командной строки и переменные окружения.
//
// Приоритет значений при инициализации:
// 1. Переменные окружения
// 2. Флаги командной строки
// 3. Значения по умолчанию (константы пакета)
package configs

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultEmptyString -  пустая строка
	DefaultEmptyString = ""
	// DefaultAddr — адрес HTTP-сервера приложения по умолчанию.
	DefaultAddr = "localhost:8080"
	// DefaultBaseURL — адрес дописываемый по умолчанию.
	DefaultBaseURL = "http://localhost:8080"
	// DefaultEnableHTTPS — отключение HTTPS по умолчанию.
	DefaultEnableHTTPS = false
	// DefaultLevelLog — уровень логирования по умолчанию.
	DefaultLevelLog = "info"
	// DefaultMigrationPath — путь к директории с миграциями базы данных по умолчанию.
	DefaultMigrationPath = "./migrations"
	// DefaultBufferSizeChan — размер буфера по умолчанию.
	DefaultBufferSizeChan = 100
	// DefaultBatchSize — размер батча по умолчанию.
	DefaultBatchSize = 10
	// DefaultFlushIntervalSecond — таймаут записи по умолчанию.
	DefaultFlushIntervalSecond = 15 * time.Second
)

// StorageConfig описывает конфигурацию для работы с хранилищем данных
type StorageConfig struct {
	// Path — путь для подключения к файловой базе
	Path string
	// PathDB — URI для подключения к базе данных
	PathDB string
	// PathMigrations — путь к файлам миграций базы данных.
	PathMigrations string
}

// LogConfig описывает конфигурацию подсистемы логирования.
type LogConfig struct {
	// Level — уровень детализации логов (например, info, debug, error).
	Level string
}

// AuthConfig описывает конфигурацию для модуля аутентификации и работы с токенами.
type AuthConfig struct {
	// JWTSecret — секретный ключ для подписи и проверки JWT-токенов.
	JWTSecret []byte
	// TokenIssuer — название издателя (issuer), указываемого в JWT-токенах.
	TokenIssuer string
}

// ServiceConfig описывает конфигурацию подсистемы сервиса.
type ServiceConfig struct {
	// BufferSizeChan содержит настройки сервиса сокращения ссылок.
	BufferSizeChan int
	// BatchSize - размер одного батча.
	BatchSize int
	// FlushInterval - периодичность очистки очереди.
	FlushInterval time.Duration
}

// AuditConfig описывает конфигурацию для модуля аудита.
type AuditConfig struct {
	// File - путь до файла аудита
	File string
	// URL - адрес удаленного сервиса аудита
	URL string
}

// ConfigJSON является главной структурой конфигурации приложения
type ConfigJSON struct {
	// Addr — сетевой адрес (хост:порт), на котором запускается HTTP-сервер приложения.
	Addr string `json:"server_address"`
	// BaseURL — сетевой адрес, который дописывается.
	BaseURL string `json:"base_url"`
	// EnableHTTPS - переменная, отвечающая за включение HTTPS
	EnableHTTPS bool `json:"enable_https"`
	// LogLevel — уровень детализации логов (например, info, debug, error).
	LogLevel string `json:"log_level"`
	// StorageFilePath — путь для подключения к файловой базе
	StorageFilePath string `json:"file_storage_path"`
	// StoragePathDB — URI для подключения к базе данных
	StoragePathDB string `json:"database_dsn"`
	// StoragePathMigrations — путь к файлам миграций базы данных.
	StoragePathMigrations string `json:"migrations_path"`
	// AuditURL - адрес удаленного сервиса аудита
	AuditURL string `json:"audit_url"`
	// AuditFile - путь до файла аудита
	AuditFile string `json:"audit_file"`
}

// Config является главной структурой конфигурации приложения
type Config struct {
	// Addr — сетевой адрес (хост:порт), на котором запускается HTTP-сервер приложения.
	Addr string
	// BaseURL — сетевой адрес, который дописывается.
	BaseURL string
	// EnableHTTPS - переменная, отвечающая за включение HTTPS
	EnableHTTPS bool
	// Log содержит настройки логирования.
	Log LogConfig
	// Storage содержит настройки подключения к хранилищу данных.
	Storage StorageConfig
	// Auth содержит настройки аутентификации.
	Auth AuthConfig
	// Service содержит настройки сервиса сокращения ссылок.
	Service ServiceConfig
	// Audit содержит настройки аудита.
	Audit AuditConfig
}

// NewConfig создает и возвращает указатель на новый экземпляр структуры Config.
func NewConfig() *Config {
	return &Config{}
}

// Init инициализирует конфигурацию.
func (c *Config) Init() {
	configFile := c.getConfigPath()
	cfg := c.parseWithJSON(configFile)

	c.parseWithFlag(cfg)

	c.Addr = c.updateEnv("SERVER_ADDRESS", c.Addr)
	c.BaseURL = c.updateEnv("BASE_URL", c.BaseURL)

	switch strings.ToLower(os.Getenv("ENABLE_HTTPS")) {
	case "true", "1", "yes", "y":
		c.EnableHTTPS = true
	default:
		c.EnableHTTPS = false
	}

	c.Log.Level = c.updateEnv("LOG_LEVEL", c.Log.Level)

	c.Storage.Path = c.updateEnv("FILE_STORAGE_PATH", c.Storage.Path)
	c.Storage.PathDB = c.updateEnv("DATABASE_DSN", c.Storage.PathDB)
	c.Storage.PathMigrations = c.updateEnv("MIGRATION_PATH", c.Storage.PathMigrations)

	c.Auth.JWTSecret = []byte(c.updateEnv("JWT_SECRET", string(c.Auth.JWTSecret)))
	c.Auth.TokenIssuer = c.updateEnv("TOKEN_ISSUER", c.Auth.TokenIssuer)

	c.Audit.File = c.updateEnv("AUDIT_FILE", c.Audit.File)
	c.Audit.URL = c.updateEnv("AUDIT_URL", c.Audit.URL)

	if bufferSizeChan, ok := os.LookupEnv("BUFFER_SIZE_CHAN"); ok {
		var errConv error
		c.Service.BufferSizeChan, errConv = strconv.Atoi(bufferSizeChan)
		if errConv != nil {
			c.Service.BufferSizeChan = DefaultBufferSizeChan
		}
	} else {
		c.Service.BufferSizeChan = DefaultBufferSizeChan
	}

	if batchSize, ok := os.LookupEnv("BATCH_SIZE"); ok {
		var errConv error
		c.Service.BatchSize, errConv = strconv.Atoi(batchSize)
		if errConv != nil {
			c.Service.BatchSize = DefaultBatchSize
		}
	} else {
		c.Service.BatchSize = DefaultBatchSize
	}

	if flushInterval, ok := os.LookupEnv("FLUSH_INTERVAL"); ok {
		temp, errConv := strconv.Atoi(flushInterval)
		if errConv != nil {
			c.Service.FlushInterval = DefaultFlushIntervalSecond
		} else {
			c.Service.FlushInterval = time.Duration(temp) * time.Second
		}
	} else {
		c.Service.FlushInterval = DefaultFlushIntervalSecond
	}
}

func (*Config) getConfigPath() string {
	if path, ok := os.LookupEnv("CONFIG"); ok {
		return path
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if strings.HasPrefix(arg, "-c=") {
			return strings.TrimPrefix(arg, "-c=")
		} else if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config=")
		} else if arg == "-c" || arg == "-config" {
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				return os.Args[i+1]
			}
		}
	}

	return DefaultEmptyString
}

func (c *Config) parseWithFlag(cfg *ConfigJSON) {
	flag.StringVar(&c.Addr, "a", cfg.Addr, "host:port")
	flag.StringVar(&c.BaseURL, "b", cfg.BaseURL, "base url")
	flag.BoolVar(&c.EnableHTTPS, "s", cfg.EnableHTTPS, "enable HTTPS")
	flag.StringVar(&c.Log.Level, "l", cfg.LogLevel, "level log")
	flag.StringVar(&c.Storage.Path, "f", cfg.StorageFilePath, "file storage path")
	flag.StringVar(&c.Storage.PathDB, "d", cfg.StoragePathDB, "file storage path")
	flag.StringVar(&c.Storage.PathMigrations, "m", cfg.StoragePathMigrations, "migrations path")
	flag.StringVar(&c.Audit.URL, "audit-url", cfg.AuditURL, "remote audit server URL")
	flag.StringVar(&c.Audit.File, "audit-file", cfg.AuditFile, "path to audit log file")

	flag.Parse()
}

func (*Config) updateEnv(name, defaultValue string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return defaultValue
}

func (*Config) parseWithJSON(path string) *ConfigJSON {
	cfg := &ConfigJSON{
		Addr:                  DefaultAddr,
		BaseURL:               DefaultBaseURL,
		EnableHTTPS:           DefaultEnableHTTPS,
		LogLevel:              DefaultLevelLog,
		StorageFilePath:       DefaultEmptyString,
		StoragePathDB:         DefaultEmptyString,
		StoragePathMigrations: DefaultMigrationPath,
		AuditFile:             DefaultEmptyString,
		AuditURL:              DefaultEmptyString,
	}
	if path == "" {
		return cfg
	}

	file, errOpen := os.Open(path)
	if errOpen != nil {
		fmt.Printf("failed to open json file: %s\n", errOpen)
		return cfg
	}

	if errUnmarshal := json.NewDecoder(file).Decode(&cfg); errUnmarshal != nil {
		fmt.Printf("failed to parse json file: %s\n", errUnmarshal)
		return cfg
	}

	return cfg
}
