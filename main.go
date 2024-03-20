package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	r := chi.NewRouter()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 認証無し
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	// 認証が必要なグループ
	r.Group(func(r chi.Router) {
		// ここに任意の認証処理を書く
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("admin page"))
		})
	})

	srv := &http.Server{
		Addr:    ":3333",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println("server error", err)
		}
	}()
	log.Println("Server is ready to handle requests at :3333")

	// graceful shutdown
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Failed to gracefully shutdown the server", err)
	}
	log.Println("Server shutdown")
}
