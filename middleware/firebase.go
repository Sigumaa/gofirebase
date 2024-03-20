package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

type firebaseContextKey struct{}

var firebaseKey = firebaseContextKey{}

type FirebaseAuthMiddleware struct {
	auth *auth.Client
}

func NewFirebaseAuthMiddleware(app *firebase.App) *FirebaseAuthMiddleware {
	auth, err := app.Auth(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return &FirebaseAuthMiddleware{auth}
}

func (m *FirebaseAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No ID token", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid ID token", http.StatusUnauthorized)
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := m.auth.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid ID token", http.StatusUnauthorized)
			return
		}

		log.Printf("Verified!")

		ctx := context.WithValue(r.Context(), firebaseKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetFirebaseToken(ctx context.Context) (*auth.Token, bool) {
	token, ok := ctx.Value(firebaseKey).(*auth.Token)
	return token, ok
}
