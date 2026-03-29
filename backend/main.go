package main

import (
	"Backend/configs"
	"Backend/handlers"
	"Backend/middlewares"
	"Backend/services"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Initialising Router")
	r := chi.NewRouter()

	// applying the common middlewares
	r.Use(middlewares.LoggingMiddleware)

	// healthcheck route
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message": "Backend Service running"}`))
	})

	// getting the configs
	serverConfig := configs.GetServerConfig()

	// creating services
	authService := services.NewAuthService()
	uploadService := services.NewUploadService()

	// creating handlers & injecting services into them
	authHandler := &handlers.AuthHandler{Service: authService}
	uploadHandler := &handlers.UploadHandler{Service: uploadService}

	// registering the routes
	r.Post(serverConfig.BackendLoginAPI, authHandler.HandleLogin)
	r.Post(serverConfig.BackendSignupAPI, authHandler.HandleSignUp)

	r.Group(func(r chi.Router) {
		// using the authorization & authentication middlewares only for these routes
		r.Use(middlewares.AuthZMiddleware)
		r.Use(middlewares.AuthNMiddleware)

		r.Post(serverConfig.BackendUploadAPI, uploadHandler.HandleReceiptUploads)
	})

	// initialising the server
	var server *http.Server

	if serverConfig.Env == "development" {
		server = &http.Server{
			Addr:         serverConfig.BackendPort,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	}

	// creating a channel to listen for OS signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // cancel the context at the end

	// starting the server in goroutine
	go func() {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Starting the Backend Service!")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		err := server.ListenAndServe()

		if err != nil {
			slog.Error(
				"Error while starting the Server:",
				slog.Any("Error:", err),
			)
		}
	}()

	// Block here and wait for the OS Background signals
	<-ctx.Done()

	// creating a context with 5 seconds timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// shutdown using the shutdown context (Attempting graceful shutdown)
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error(
			"Server forced to shutdown:",
			slog.Any("Error", err),
		)
	}

	slog.Info("Server Exited!")
}
