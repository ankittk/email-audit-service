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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ankittk/email-audit-service/internal/config"
	"github.com/ankittk/email-audit-service/internal/handler"
	"github.com/ankittk/email-audit-service/internal/middleware"
	pb "github.com/ankittk/email-audit-service/proto"
)

func main() {
	grpcConn, err := grpc.NewClient(config.RulesEngineAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Rules Engine: %v", err)
	}
	defer grpcConn.Close()

	h := &handler.Handler{RulesClient: pb.NewRulesEngineServiceClient(grpcConn)}
	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware, middleware.CORSMiddleware, middleware.ErrorRecoveryMiddleware)
	h.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("🚀 HTTP server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
		if err := srv.Close(); err != nil {
			log.Fatalf("Close error: %v", err)
		}
	}

	log.Println("HTTP Server gracefully stopped")
}
