package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var mc *memcache.Client

func init() {
	mc = memcache.New("localhost:11211")
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserSession struct {
	Username     string   `json:"username"`
	Role         string   `json:"role"`
	AllowedPages []string `json:"allowedPages"`
	Token        string   `json:"token"`
}

type AuthResponse struct {
	Username     string   `json:"username"`
	Role         string   `json:"role"`
	AllowedPages []string `json:"allowedPages"`
}

func hashSHA256(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

func generateAccessToken() string {
	nanoTime := strconv.FormatInt(time.Now().UnixNano(), 10)
	token := hashSHA256(nanoTime)
	return token
}

func resourceEndpoint(w http.ResponseWriter, r *http.Request) {
	userDetails, ok := r.Context().Value("userDetails").(AuthResponse)
	if !ok {
		httpError(w, http.StatusInternalServerError, "Invalid session data")
		return
	}
	respondWithJSON(w, http.StatusOK, userDetails)
}

func httpError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func loginEndpoint(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		httpError(w, http.StatusBadRequest, "Unable to parse JSON")
		return
	}

	if authenticate(user) {
		token, session := createSession(user)
		storeSession(user.Username, session)
		respondWithJSON(w, http.StatusOK, map[string]string{"token": token})
	} else {
		httpError(w, http.StatusUnauthorized, "Invalid credentials")
	}
}

func authenticate(user User) bool {
	return user.Username == "john" && user.Password == "pwd123"
}

func createSession(user User) (string, UserSession) {
	token := generateAccessToken()
	session := UserSession{
		Username:     user.Username,
		Role:         "admin",
		AllowedPages: []string{"/dashboard", "/settings"},
		Token:        hashSHA256(token),
	}
	return token, session
}

func storeSession(username string, session UserSession) {
	data, _ := json.Marshal(session)
	mc.Set(&memcache.Item{Key: hashSHA256(username), Value: data, Expiration: 1800})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-AUTH")
		username := r.Header.Get("X-USER")

		session, err := retrieveSession(username)
		if err != nil {
			httpError(w, http.StatusUnauthorized, "Session expired or not found")
			return
		}

		if session.Token != hashSHA256(token) {
			httpError(w, http.StatusUnauthorized, "Invalid session or token")
			return
		}

		authResponse := AuthResponse{
			Username:     session.Username,
			Role:         session.Role,
			AllowedPages: session.AllowedPages,
		}
		ctx := context.WithValue(r.Context(), "userDetails", authResponse)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func retrieveSession(username string) (UserSession, error) {
	data, err := mc.Get(hashSHA256(username))
	if err != nil {
		return UserSession{}, err
	}
	var session UserSession
	json.Unmarshal(data.Value, &session)
	return session, nil
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/api/login", loginEndpoint)
	router.With(AuthMiddleware).Get("/api/resource", resourceEndpoint)
	fmt.Println("Starting server...")
	http.ListenAndServe(":8080", router)
}
