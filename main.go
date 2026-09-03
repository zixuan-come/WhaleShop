package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zixuan-come/whaleshop/internal/handler"
	"github.com/zixuan-come/whaleshop/internal/middleware"
	"github.com/zixuan-come/whaleshop/internal/store"
)

func main() {
	st := store.New()
	oh := handler.NewOrders(st)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle("GET /orders", middleware.Project(http.HandlerFunc(oh.List)))
	mux.Handle("POST /orders", middleware.Project(http.HandlerFunc(oh.Create)))
	mux.Handle("GET /orders/slow", middleware.Project(http.HandlerFunc(oh.Slow)))
	mux.Handle("GET /orders/error", middleware.Project(http.HandlerFunc(oh.Error)))
	mux.Handle("GET /orders/{id}", middleware.Project(http.HandlerFunc(oh.Get)))
	mux.Handle("PUT /orders/{id}", middleware.Project(http.HandlerFunc(oh.Update)))
	mux.Handle("DELETE /orders/{id}", middleware.Project(http.HandlerFunc(oh.Delete)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("whaleshop listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
