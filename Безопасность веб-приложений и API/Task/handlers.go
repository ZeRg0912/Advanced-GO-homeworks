package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// RegisterHandler обрабатывает регистрацию нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := parseJSONRequest(r, &req); err != nil {
		sendErrorResponse(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if err := validateRegisterRequest(&req); err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	exists, err := UserExistsByEmail(req.Email)
	if err != nil {
		log.Printf("Error checking user existence: %v", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		sendErrorResponse(w, "User with this email already exists", http.StatusConflict)
		return
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user, err := CreateUser(req.Email, req.Username, passwordHash)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			sendErrorResponse(w, "User with this email or username already exists", http.StatusConflict)
			return
		}
		log.Printf("Error creating user: %v", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(*user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, AuthResponse{Token: token, User: *user}, http.StatusCreated)
}

// LoginHandler обрабатывает вход пользователя
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := parseJSONRequest(r, &req); err != nil {
		sendErrorResponse(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if err := validateLoginRequest(&req); err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorResponse(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		log.Printf("Error getting user by email: %v", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !CheckPassword(req.Password, user.PasswordHash) {
		sendErrorResponse(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(*user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, AuthResponse{Token: token, User: *user}, http.StatusOK)
}

// ProfileHandler возвращает профиль текущего пользователя
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := GetUserIDFromContext(r)
	if !ok {
		sendErrorResponse(w, "User context is missing", http.StatusUnauthorized)
		return
	}

	user, err := GetUserByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendErrorResponse(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user by id: %v", err)
		sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, user, http.StatusOK)
}

// HealthHandler проверяет состояние сервиса
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем подключение к БД
	if db != nil {
		if err := db.Ping(); err != nil {
			sendErrorResponse(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}
	}

	// Возвращаем статус OK
	response := map[string]string{
		"status":  "ok",
		"message": "Service is running",
	}
	sendJSONResponse(w, response, http.StatusOK)
}

// sendJSONResponse отправляет JSON ответ (вспомогательная функция)
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendErrorResponse отправляет JSON ответ с ошибкой (вспомогательная функция)
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}

// parseJSONRequest парсит JSON из тела запроса (вспомогательная функция)
func parseJSONRequest(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Строгая проверка полей

	return decoder.Decode(v)
}

// validateRegisterRequest валидирует данные регистрации
func validateRegisterRequest(req *RegisterRequest) error {
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	if len(req.Username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(req.Username) > 30 {
		return fmt.Errorf("username must be at most 30 characters long")
	}
	for _, ch := range req.Username {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			continue
		}
		return fmt.Errorf("username can contain only letters, digits, underscore and hyphen")
	}

	return nil
}

// validateLoginRequest валидирует данные входа
func validateLoginRequest(req *LoginRequest) error {
	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
