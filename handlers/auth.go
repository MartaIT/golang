package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"golang/middleware"
	"golang/store"
	"golang/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Store *store.MemStore
}

// RegisterPayload simple
type RegisterPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var p RegisterPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if p.Username == "" || p.Password == "" {
		http.Error(w, "username & password required", http.StatusBadRequest)
		return
	}
	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	role := p.Role
	if role == "" {
		role = "user"
	}
	_, err = h.Store.CreateUser(p.Username, string(hashed), role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("user created"))
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var p LoginPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	user, err := h.Store.GetByUsername(p.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(p.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := utils.GenerateToken(user.Username, user.Role, time.Hour*1) // 1 hour TTL
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}
	resp := map[string]string{"token": token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Protected endpoint example
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	claims, ok := v.(*utils.Claims)
	if !ok {
		http.Error(w, "invalid auth context", http.StatusUnauthorized)
		return
	}
	resp := map[string]string{
		"username": claims.Username,
		"role":     claims.Role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func NewAuthHandler(store *store.MemStore) *AuthHandler {
	return &AuthHandler{Store: store}
}
