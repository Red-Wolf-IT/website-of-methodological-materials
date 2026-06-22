package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"website-of-methodological-materials/internal/config"
	"website-of-methodological-materials/internal/db"
	"website-of-methodological-materials/internal/handlers"
	"website-of-methodological-materials/internal/repository/postgres"
	"website-of-methodological-materials/internal/server"
	"website-of-methodological-materials/internal/service"
	"website-of-methodological-materials/internal/validator"
)

func main() {
	// .env подхватывается только локально; в проде переменные задаются снаружи
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// собираем цепочку: repo → service → handler
	healthRepo := postgres.NewHealthRepository(pool)
	healthService := service.NewHealthService(healthRepo)
	healthHandler := handlers.NewHealthHandler(healthService)

	manualRepo := postgres.NewManualRepository(pool)
	manualService := service.NewManualService(manualRepo)
	manualHandler := handlers.NewManualHandler(manualService, validator.New())

	httpServer := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: server.New(healthHandler, manualHandler),
	}

	go func() {
		log.Printf("server listening on %s", cfg.ServerAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// ждём Ctrl+C или SIGTERM и выключаем сервер без обрыва запросов
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}

	log.Println("server stopped")
}
