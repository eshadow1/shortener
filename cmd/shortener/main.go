package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eshadow1/shortener/internal/configs"
	"github.com/eshadow1/shortener/internal/handler"
	"github.com/eshadow1/shortener/internal/repository"
	"github.com/eshadow1/shortener/internal/service"

	"net/http"
)

const (
	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 30 * time.Second
)

func main() {
	cfg := configs.NewConfig()
	cfg.ParseWithFlag()

	r := repository.NewMemoryRepository()
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
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return
	}
}
