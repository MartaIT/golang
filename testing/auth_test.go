package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang/handlers"
	"golang/middleware"
	"golang/store"
	"golang/utils"
)

func setupServer() http.Handler {
	st := store.NewStore()
	ah := handlers.NewAuthHandler(st)

	mux := http.NewServeMux()
	mux.HandleFunc("/register", ah.Register)
	mux.HandleFunc("/login", ah.Login)
	mux.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(ah.Profile)))

	return mux
}

func TestRegisterLoginProfile(t *testing.T) {
	srv := setupServer()

	// REGISTER
	payload := map[string]string{"username": "bob", "password": "bobpass"}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 created got %d", rr.Code)
	}

	// LOGIN
	req = httptest.NewRequest("POST", "/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 login got %d", rr.Code)
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	token := resp["token"]
	if token == "" {
		t.Fatal("expected token")
	}

	// PROFILE
	req = httptest.NewRequest("GET", "/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 profile got %d", rr.Code)
	}

	// Validate profile response
	var prof map[string]string
	json.NewDecoder(rr.Body).Decode(&prof)

	if prof["username"] != "bob" {
		t.Fatalf("expected username bob got %s", prof["username"])
	}

	// Validate token claims
	claims, err := utils.ParseToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if claims.Username != "bob" {
		t.Fatalf("expected claims.Username = bob got %s", claims.Username)
	}
}
