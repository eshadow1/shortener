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
	// DefaultGRPCAddr — адрес HTTP-сервера приложения по умолчанию.
	DefaultGRPCAddr = ":50051"
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
	// DefaultTLSCertFile - дефолтный сертификат для HTTPS
	DefaultTLSCertFile = "cert/cert.pem"
	// DefaultTLSKeyFile  - дефолтный ключ для HTTPS
	DefaultTLSKeyFile = "cert/key.pem"
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

// HTTPSConfig описывает конфигурацию для HTTPS.
type HTTPSConfig struct {
	// EnableHTTPS - переменная, отвечающая за включение HTTPS
	EnableHTTPS bool
	// TLSCertFile - сертификат для HTTPS
	TLSCertFile string
	// TLSKeyFile  - ключ для HTTPS
	TLSKeyFile string
}

// ConfigJSON является главной структурой конфигурации приложения
type ConfigJSON struct {
	// Addr — сетевой адрес (хост:порт), на котором запускается HTTP-сервер приложения.
	Addr string `json:"server_address"`
	// GRPCAddr — сетевой адрес для gRPC, который дописывается.
	GRPCAddr string `json:"grpc_address"`
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
	// TrustedSubnet - строковое представление бесклассовой адресации (CIDR)
	TrustedSubnet string `json:"trusted_subnet"`
}

// Config является главной структурой конфигурации приложения
type Config struct {
	// Addr — сетевой адрес (хост:порт), на котором запускается HTTP-сервер приложения.
	Addr string
	// GRPCAddr — сетевой адрес для gRPC, который дописывается.
	GRPCAddr string
	// BaseURL — сетевой адрес, который дописывается.
	BaseURL string
	// TrustedSubnet - строковое представление бесклассовой адресации (CIDR)
	TrustedSubnet string
	// HTTPS содержит настройки HTTPS
	HTTPS HTTPSConfig
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
	cfg, errParse := c.parseWithJSON(configFile)
	if errParse != nil {
		fmt.Fprintf(os.Stderr, "failed to parse file: %s\n", errParse)
	}

	c.parseWithFlag(cfg)

	c.Addr = c.updateEnv("SERVER_ADDRESS", c.Addr)
	c.GRPCAddr = c.updateEnv("GRPC_ADDRESS", c.GRPCAddr)
	c.BaseURL = c.updateEnv("BASE_URL", c.BaseURL)

	c.TrustedSubnet = c.updateEnv("TRUSTED_SUBNET", c.TrustedSubnet)

	enableHTTPS, errParseBool := strconv.ParseBool(os.Getenv("ENABLE_HTTPS"))
	if errParseBool != nil {
		c.HTTPS.EnableHTTPS = false
	} else {
		c.HTTPS.EnableHTTPS = enableHTTPS
	}
	c.HTTPS.TLSKeyFile = c.updateEnv("TLS_KEY_FILE", DefaultTLSKeyFile)
	c.HTTPS.TLSCertFile = c.updateEnv("TLS_CERT_FILE", DefaultTLSCertFile)

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
	flag.StringVar(&c.GRPCAddr, "g", cfg.GRPCAddr, "host:port")
	flag.StringVar(&c.BaseURL, "b", cfg.BaseURL, "base url")
	flag.StringVar(&c.TrustedSubnet, "t", cfg.TrustedSubnet, "trusted subnet")
	flag.BoolVar(&c.HTTPS.EnableHTTPS, "s", cfg.EnableHTTPS, "enable HTTPS")
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

func (*Config) parseWithJSON(path string) (*ConfigJSON, error) {
	cfg := &ConfigJSON{
		Addr:                  DefaultAddr,
		GRPCAddr:              DefaultGRPCAddr,
		BaseURL:               DefaultBaseURL,
		EnableHTTPS:           DefaultEnableHTTPS,
		LogLevel:              DefaultLevelLog,
		StorageFilePath:       DefaultEmptyString,
		StoragePathDB:         DefaultEmptyString,
		StoragePathMigrations: DefaultMigrationPath,
		AuditFile:             DefaultEmptyString,
		AuditURL:              DefaultEmptyString,
		TrustedSubnet:         DefaultEmptyString,
	}
	if path == "" {
		return cfg, nil
	}

	file, errOpen := os.Open(path)
	if errOpen != nil {
		return cfg, fmt.Errorf("failed to open json file: %w", errOpen)
	}
	defer file.Close()

	if errUnmarshal := json.NewDecoder(file).Decode(&cfg); errUnmarshal != nil {
		return cfg, fmt.Errorf("failed to parse json file: %w", errUnmarshal)
	}

	return cfg, nil
}
