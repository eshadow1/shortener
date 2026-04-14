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

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/loggers"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"
)

const (
	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 30 * time.Second
)

func main() {
	cfg := configs.NewConfig()
	cfg.Init()

	errCreateLog := loggers.CreateLogger(cfg.Log.Level)
	if errCreateLog != nil {
		fmt.Println("Error creating logger:", errCreateLog)
		return
	}

	r := repository.NewMemoryRepository(cfg.Storage.Path)
	defer r.Close()
	s := service.NewShortenerService(r)
	h := handler.NewHandler(cfg, s)

	rs := handler.InitRouter(h)

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      rs,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			loggers.Log.Fatalf("Server failed: %v", err)
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
