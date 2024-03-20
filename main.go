package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mw "github.com/Sigumaa/gofirebase/middleware"

	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if os.Getenv("FB_SECRET_CREDENTIAL") == "" {
		log.Fatal("FB_SECRET_CREDENTIAL must be set")
	}
}

func main() {
	r := chi.NewRouter()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(os.Getenv("FB_SECRET_CREDENTIAL"))))
	if err != nil {
		log.Fatal(err)
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 認証無し
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	firebaseAuth := mw.NewFirebaseAuthMiddleware(app)

	// 認証が必要なグループ
	r.Group(func(r chi.Router) {
		r.Use(firebaseAuth.Middleware)
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			log.Println(mw.GetFirebaseToken(r.Context()))
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Failed to gracefully shutdown the server", err)
	}
	log.Println("Server shutdown")
}
