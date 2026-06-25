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
	"website-of-methodological-materials/internal/storage"
	"website-of-methodological-materials/internal/validator"
	"website-of-methodological-materials/internal/worker"
)

func main() {
	// .env подхватывается только локально; в проде переменные задаются снаружи
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.AdminToken == "" {
		log.Fatal("ADMIN_TOKEN is required")
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	fileStorage, err := storage.NewFileStorage(cfg.StorageDir)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	// собираем цепочку: repo → service → handler
	healthRepo := postgres.NewHealthRepository(pool)
	healthService := service.NewHealthService(healthRepo)
	healthHandler := handlers.NewHealthHandler(healthService)

	manualRepo := postgres.NewManualRepository(pool)
	manualService := service.NewManualService(manualRepo, fileStorage)

	viewsWorker := worker.NewViewsWorker(manualRepo, 100)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go viewsWorker.Run(workerCtx)

	v := validator.New()
	manualHandler := handlers.NewManualHandler(manualService, v, viewsWorker.ViewsChan())

	tagRepo := postgres.NewTagRepository(pool)
	tagService := service.NewTagService(tagRepo)
	tagHandler := handlers.NewTagHandler(tagService, v)
	fileHandler := handlers.NewFileHandler(manualService)

	httpServer := &http.Server{
		Addr: cfg.ServerAddr,
		Handler: server.New(
			server.Config{AdminToken: cfg.AdminToken},
			healthHandler,
			manualHandler,
			tagHandler,
			fileHandler,
		),
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

	workerCancel()
	viewsWorker.Wait()

	log.Println("server stopped")
}
