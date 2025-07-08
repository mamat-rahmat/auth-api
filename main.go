package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type RegisterRequest struct {
	NIK  string `json:"nik"`
	Role string `json:"role"`
}

type RegisterResponse struct {
	ID       int    `json:"id"`
	NIK      string `json:"nik"`
	Role     string `json:"role"`
	Password string `json:"password"`
	Message  string `json:"message"`
}

type LoginRequest struct {
	NIK      string `json:"nik"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID          int    `json:"id"`
	NIK         string `json:"nik"`
	Role        string `json:"role"`
	AccessToken string `json:"access_token"`
	Message     string `json:"message"`
}

type ProfileResponse struct {
	ID      int    `json:"id"`
	NIK     string `json:"nik"`
	Role    string `json:"role"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type User struct {
	ID       int    `json:"id"`
	NIK      string `json:"nik"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

type Claims struct {
	ID   int    `json:"id"`
	NIK  string `json:"nik"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

var users = make(map[string]User)
var userCounter = 1

var jwtSecret []byte

func generatePassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, 6)

	for i := range password {
		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[randomIndex.Int64()]
	}

	return string(password)
}

func generateJWT(user User) (string, error) {
	claims := &Claims{
		ID:   user.ID,
		NIK:  user.NIK,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "Token tidak ditemukan",
			})
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "Format token tidak valid",
			})
			return
		}

		tokenString := tokenParts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "Token tidak valid",
			})
			return
		}

		// Tambahkan claims ke context request
		r.Header.Set("X-User-ID", strconv.Itoa(claims.ID))
		r.Header.Set("X-User-NIK", claims.NIK)
		r.Header.Set("X-User-Role", claims.Role)

		next.ServeHTTP(w, r)
	}
}

// Endpoint 1: Register user dan generate password
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Hanya menerima POST request",
		})
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Format JSON tidak valid",
		})
		return
	}

	if req.NIK == "" || req.Role == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "NIK dan Role harus diisi",
		})
		return
	}

	if _, exists := users[req.NIK]; exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "conflict",
			Message: "NIK sudah terdaftar",
		})
		return
	}

	password := generatePassword()

	user := User{
		ID:       userCounter,
		NIK:      req.NIK,
		Role:     req.Role,
		Password: password,
	}
	users[req.NIK] = user
	userCounter++

	response := RegisterResponse{
		ID:       user.ID,
		NIK:      user.NIK,
		Role:     user.Role,
		Password: password,
		Message:  "User berhasil didaftarkan",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Endpoint 2: Login dan generate JWT token
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Hanya menerima POST request",
		})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "Format JSON tidak valid",
		})
		return
	}

	if req.NIK == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "bad_request",
			Message: "NIK dan Password harus diisi",
		})
		return
	}

	user, exists := users[req.NIK]
	if !exists || user.Password != req.Password {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "unauthorized",
			Message: "NIK atau Password salah",
		})
		return
	}

	token, err := generateJWT(user)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "internal_error",
			Message: "Gagal membuat token",
		})
		return
	}

	response := LoginResponse{
		ID:          user.ID,
		NIK:         user.NIK,
		Role:        user.Role,
		AccessToken: token,
		Message:     "Login berhasil",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Endpoint 3: Get profile dengan JWT validation
func profileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Hanya menerima GET request",
		})
		return
	}

	userID := r.Header.Get("X-User-ID")
	userNIK := r.Header.Get("X-User-NIK")
	userRole := r.Header.Get("X-User-Role")

	id, _ := strconv.Atoi(userID)

	response := ProfileResponse{
		ID:      id,
		NIK:     userNIK,
		Role:    userRole,
		Message: "Data profile berhasil diambil",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Auth API is running",
	})
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY environment variable is required")
	}

	jwtSecret = []byte(secretKey)
	log.Println("Environment variables loaded successfully")
}

func main() {
	loadEnv()

	r := mux.NewRouter()

	// Routes
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/register", registerHandler).Methods("POST")
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/profile", validateJWT(profileHandler)).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on :%s...\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("- GET  /health")
	fmt.Println("- POST /api/register")
	fmt.Println("- POST /api/login")
	fmt.Println("- GET  /api/profile (requires JWT)")

	log.Fatal(http.ListenAndServe(":"+port, r))
}
