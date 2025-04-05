package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"bitterlink/core/internal/agency"
	"bitterlink/core/internal/config"
	"bitterlink/core/internal/db"
	"bitterlink/core/internal/logging"
	"bitterlink/core/internal/repository"
	"bitterlink/core/internal/transport/http"
	"bitterlink/core/internal/worker"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logging.SetupLogging()
	config.LoadEnv()
	log.Println("INFO: Starting application...")

	databasePool, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("FATAL: Database initialization failed: %v", err)
	}
	log.Println("INFO: Database connection ready.")

	// --- Timeout Checker Worker ---
	// Configuration (Read from Env Vars or defaults)
	pollIntervalSeconds, _ := strconv.Atoi(os.Getenv("CHECKER_POLL_INTERVAL_SECONDS"))
	if pollIntervalSeconds <= 0 {
		pollIntervalSeconds = 30
	}
	batchSize, _ := strconv.Atoi(os.Getenv("CHECKER_BATCH_SIZE"))
	if batchSize <= 0 {
		batchSize = 10
	}
	checkerConfig := worker.Config{
		PollInterval: time.Duration(pollIntervalSeconds) * time.Second,
		BatchSize:    batchSize,
	}
	timeoutChecker := worker.NewTimeoutChecker(databasePool, checkerConfig)

	// Create a context that can be cancelled for graceful shutdown
	// Link it to SIGINT/SIGTERM signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the checker worker in a separate goroutine
	// Pass the cancellable context
	go timeoutChecker.Start(ctx)

	// Create repository instances
	checkRepo := repository.NewMySQLCheckRepository(databasePool)
	// userRepo := repository.NewMySQLUserRepository(dbPool) // etc.

	// Create handler instances, injecting dependencies
	pingHandler := httptransport.NewPingHandler(checkRepo)
	checkHandler := httptransport.NewCheckHandler(checkRepo)
	// checkHandler := httptransport.NewCheckHandler(checkRepo) // For API CRUD

	router := gin.Default()

	httptransport.RegisterRoutes(router, pingHandler, checkHandler, databasePool, checkRepo)
	log.Println("INFO: HTTP routes registered.")

	srvPort := os.Getenv("SERVER_PORT")
	if srvPort == "" {
		srvPort = "8080"
	}

	if !agency.IsNumeric(srvPort) {
		log.Printf("ERROR: Server port: %s is not numeric.\n", srvPort)
	}

	srv := &http.Server{
		Addr:    ":" + srvPort,
		Handler: router,
		// Add Read/Write timeouts for production readiness
		// ReadTimeout: 5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		// IdleTimeout: 120 * time.Second,
	}

	go func() {
		log.Printf("INFO: Starting HTTP server on port :%s", srvPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("FATAL: listen: %s\n", err)
		}
	}()

	// --- Graceful Shutdown ---
	// Wait for interrupt signal (captured by signal.NotifyContext)
	<-ctx.Done()

	stop()
	log.Println("INFO: Shutting down server and workers...")

	// Create a deadline context for the shutdown process.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout slightly
	defer cancel()

	// Attempt to gracefully shut down the HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("WARN: Server shutdown failed: %v", err)
	} else {
		log.Println("INFO: Server gracefully stopped.")
	}

	// At this point, the context passed to timeoutChecker.Start() is cancelled,
	// so its loop should exit cleanly. You might add a WaitGroup if you
	// need to explicitly wait for background workers like the checker to finish.
	log.Println("INFO: Application exited.")
}
