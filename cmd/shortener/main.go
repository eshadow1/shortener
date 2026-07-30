// Package main является точкой входа приложения Shortener
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/eshadow1/shortener/internal/audit"
	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"
)

const (
	// defaultReadTimeout — максимальное время чтения всего HTTP-запроса,
	// включая тело.
	defaultReadTimeout = 15 * time.Second
	// defaultWriteTimeout — максимальное время записи HTTP-ответа.
	defaultWriteTimeout = 15 * time.Second
	// defaultIdleTimeout — максимальное время простоя keep-alive соединения.
	defaultIdleTimeout = 60 * time.Second
	// defaultShutdownTimeout — максимальное время, отводимое на graceful shutdown
	// сервера и фоновых процессов.
	defaultShutdownTimeout = 30 * time.Second
	// defaultVersionValue - дефолтное значение для формирования информации о версии
	defaultVersionValue = "N/A"
)

var (
	buildVersion = defaultVersionValue
	buildDate    = defaultVersionValue
	buildCommit  = defaultVersionValue
)

// main — точка входа приложения. Выполняет инициализацию всех компонентов,
// запуск HTTP-сервера и фонового воркера, ожидание сигнала завершения
// и graceful shutdown.
func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit)

	cfg := configs.NewConfig()
	cfg.Init()

	errCreateLog := loggers.CreateLogger(cfg.Log.Level)
	if errCreateLog != nil {
		fmt.Println("Error creating logger:", errCreateLog)
		return
	}

	var r service.Repository
	var rc service.RepoChecker
	if cfg.Storage.PathDB != "" {
		pdb, errCreate := repository.NewPostgreSQLRepository(cfg.Storage)
		if errCreate != nil {
			loggers.Log.Errorf("error creating connection db: %v", errCreate)
			return
		}
		r = pdb
		rc = pdb
	} else {
		r = repository.NewMemoryRepository(cfg.Storage.Path)
	}
	defer r.Close()

	a := service.NewAuditBroker()
	defer a.Close()

	if af := audit.NewFileObserver(cfg.Audit.File); af != nil {
		a.Register(af)
	}

	if ar := audit.NewRemoteObserver(cfg.Audit.URL); ar != nil {
		a.Register(ar)
	}

	s := service.NewShortenerService(r, cfg.Service)
	defer s.Close()
	c := service.NewCheckerService(rc)
	h := handler.NewHandler(cfg, s, c)

	rs := handler.InitRouter(cfg, h, a)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      rs,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if cfg.HTTPS.EnableHTTPS {
			loggers.Log.Infof("Starting HTTPS server on %s (cert: %s, key: %s)", cfg.Addr, cfg.HTTPS.TLSCertFile, cfg.HTTPS.TLSKeyFile)
			errTLS := server.ListenAndServeTLS(cfg.HTTPS.TLSCertFile, cfg.HTTPS.TLSKeyFile)
			if errTLS != nil && !errors.Is(errTLS, http.ErrServerClosed) {
				loggers.Log.Fatalf("HTTPS server failed: %v", errTLS)
			}
		} else {
			loggers.Log.Infof("Starting HTTP server on %s", cfg.Addr)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				loggers.Log.Fatalf("Server failed: %v", err)
			}
		}
	}()

	<-quit
	loggers.Log.Infoln("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		loggers.Log.Infof("Server forced to shutdown: %v", err)
		return
	}
}
