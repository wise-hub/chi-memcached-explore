package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var mc *memcache.Client

func init() {
	mc = memcache.New("memcache:11211")
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthMetadata struct {
	Username     string   `json:"username"`
	UserId       string   `json:"userId"`
	Role         string   `json:"role"`
	AllowedPages []string `json:"allowedPages"`
}

func hashSHA256(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

func resourceEndpoint(w http.ResponseWriter, r *http.Request) {
	userDetails, ok := r.Context().Value("userDetails").(AuthMetadata)
	if !ok {
		httpError(w, http.StatusInternalServerError, "Invalid session data")
		return
	}
	respondWithJSON(w, http.StatusOK, userDetails)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func httpError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func authenticate(user User) (bool, AuthMetadata) {

	// mock authentication service
	if user.Username == "john" && user.Password == "pwd123" {
		authMedatada := AuthMetadata{
			Username:     user.Username,
			UserId:       "123456",
			Role:         "admin",
			AllowedPages: []string{"/dashboard", "/settings"},
		}
		return true, authMedatada
	}
	return false, AuthMetadata{}
}

func loginEndpoint(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		httpError(w, http.StatusBadRequest, "Unable to parse JSON")
		return
	}

	authSuccess, authMetadata := authenticate(user)
	if !authSuccess {
		httpError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token := hashSHA256(strconv.FormatInt(time.Now().UnixNano(), 10))

	allowedPages := strings.Join(authMetadata.AllowedPages, ",")
	serializedSession := fmt.Sprintf("%s|%s|%s|%s|%s", authMetadata.Username, authMetadata.UserId, authMetadata.Role, allowedPages, hashSHA256(token))

	err = mc.Set(&memcache.Item{
		Key:        authMetadata.UserId,
		Value:      []byte(serializedSession),
		Expiration: 60 * 30,
	})
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed to connect to Memcache server")
		fmt.Println(err)
		return
	}

	encryptedUserId, err := Encrypt(authMetadata.UserId)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed to encrypt user data")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"accessToken": encryptedUserId + "." + token})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenParts := strings.Split(string(r.Header.Get("X-ACCESS-TOKEN")), ".")
		if len(tokenParts) != 2 {
			httpError(w, http.StatusUnauthorized, "Invalid token data")
			return
		}

		decryptedUserId, err := Decrypt(tokenParts[0])
		if err != nil {
			httpError(w, http.StatusUnauthorized, "Failed to decrypt user data")
			return
		}

		item, err := mc.Get(decryptedUserId)
		if err != nil {
			httpError(w, http.StatusUnauthorized, "Session expired or not found")
			return
		}

		parts := strings.Split(string(item.Value), "|")
		if len(parts) != 5 {
			httpError(w, http.StatusUnauthorized, "Invalid session data")
			return
		}

		if parts[4] != hashSHA256(tokenParts[1]) {
			httpError(w, http.StatusUnauthorized, "Invalid session or token")
			return
		}

		allowedPages := strings.Split(parts[3], ",")

		authMetadata := AuthMetadata{
			Username:     parts[0],
			UserId:       parts[1],
			Role:         parts[2],
			AllowedPages: allowedPages,
		}

		ctx := context.WithValue(r.Context(), "userDetails", authMetadata)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/api/login", loginEndpoint)
	router.With(authMiddleware).Get("/api/resource", resourceEndpoint)
	fmt.Println("Starting server...")
	http.ListenAndServe(":8080", router)
}
