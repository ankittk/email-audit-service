package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/ankittk/email-audit-service/internal/handler"
	"github.com/ankittk/email-audit-service/internal/middleware"
)

func main() {
	r := mux.NewRouter()

	// Add middleware
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.ErrorRecoveryMiddleware)

	// Routes
	r.HandleFunc("/health", handler.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/upload-email", handler.UploadEmailHandler).Methods("POST")
	r.HandleFunc("/audit/{audit_id}", handler.GetAuditReportHandler).Methods("GET")
	r.HandleFunc("/audit/summary", handler.GetAuditSummaryHandler).Methods("GET")
	r.HandleFunc("/rules", handler.ListRulesHandler).Methods("GET")
	r.HandleFunc("/rules", handler.AddRuleHandler).Methods("POST")
	r.HandleFunc("/rules/{rule_id}", handler.UpdateRuleHandler).Methods("PUT")
	r.HandleFunc("/rules/{rule_id}", handler.DeleteRuleHandler).Methods("DELETE")

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Println("🚀 HTTP server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown error: %v", err)
		if err := srv.Close(); err != nil {
			log.Fatalf("HTTP server Close error: %v", err)
		}
	}

	log.Println("HTTP Server gracefully stopped")
}
