package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret-key-demo")

// ---------------- STORE ----------------
type Store struct {
	users map[string]string
}

func NewStore() *Store {
	return &Store{users: make(map[string]string)}
}

func (s *Store) CreateUser(username, password string) error {
	s.users[username] = password
	return nil
}

func (s *Store) ValidateUser(username, password string) bool {
	pass, ok := s.users[username]
	return ok && pass == password
}

// ---------------- HANDLER ----------------

type AuthHandler struct {
	Store *Store
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var u map[string]string
	json.NewDecoder(r.Body).Decode(&u)

	username := u["username"]
	password := u["password"]

	a.Store.CreateUser(username, password)

	w.WriteHeader(http.StatusCreated)
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var u map[string]string
	json.NewDecoder(r.Body).Decode(&u)

	username := u["username"]
	password := u["password"]

	if !a.Store.ValidateUser(username, password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	tokenStr, _ := token.SignedString(jwtKey)

	resp := map[string]string{"token": tokenStr}
	json.NewEncoder(w).Encode(resp)
}

func (a *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(ctxKey("user")).(string)
	resp := map[string]string{"username": username}
	json.NewEncoder(w).Encode(resp)
}

// ---------------- MIDDLEWARE ----------------

type ctxKey string

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		username := claims["sub"].(string)

		// FIX: simpan username langsung ke context
		ctx := context.WithValue(r.Context(), ctxKey("user"), username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
