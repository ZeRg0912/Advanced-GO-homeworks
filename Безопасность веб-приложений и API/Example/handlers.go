package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// RegisterHandler обрабатывает регистрацию нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Реализуйте регистрацию пользователя
	//
	// Пошаговый план:
	// 1. Распарсите JSON из тела запроса в структуру RegisterRequest
	// 2. Проведите валидацию данных (email, username, password)
	// 3. Проверьте, что пользователь с таким email не существует
	// 4. Захешируйте пароль с помощью функции HashPassword()
	// 5. Создайте пользователя в БД с помощью CreateUser()
	// 6. Сгенерируйте JWT токен с помощью GenerateToken()
	// 7. Верните ответ с токеном и данными пользователя
	//
	// Подсказки:
	// - Используйте json.NewDecoder(r.Body).Decode() для парсинга JSON
	// - Проверьте что все обязательные поля заполнены
	// - При ошибках возвращайте соответствующие HTTP статусы
	// - 400 для невалидных данных, 409 для дубликатов, 500 для внутренних ошибок
	// - Не забудьте установить Content-Type: application/json для ответа

	http.Error(w, "Registration not implemented", http.StatusNotImplemented)
}

// LoginHandler обрабатывает вход пользователя
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Реализуйте авторизацию пользователя
	//
	// Пошаговый план:
	// 1. Распарсите JSON из тела запроса в структуру LoginRequest
	// 2. Проведите базовую валидацию (email и password не пустые)
	// 3. Найдите пользователя по email с помощью GetUserByEmail()
	// 4. Проверьте пароль с помощью CheckPassword()
	// 5. Сгенерируйте JWT токен с помощью GenerateToken()
	// 6. Верните ответ с токеном и данными пользователя
	//
	// Важные моменты безопасности:
	// - При неверном email или пароле возвращайте одинаковое сообщение
	//   "Invalid email or password" чтобы не раскрывать существование email
	// - Используйте HTTP статус 401 для неверных учетных данных
	// - Не возвращайте password_hash в ответе

	http.Error(w, "Login not implemented", http.StatusNotImplemented)
}

// ProfileHandler возвращает профиль текущего пользователя
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Реализуйте получение профиля пользователя
	//
	// Пошаговый план:
	// 1. Получите ID пользователя из контекста с помощью GetUserIDFromContext()
	// 2. Загрузите данные пользователя из БД с помощью GetUserByID()
	// 3. Верните данные пользователя в JSON формате
	//
	// Примечания:
	// - Этот обработчик вызывается только после AuthMiddleware
	// - Контекст уже должен содержать userID
	// - Если пользователь не найден - верните 404
	// - Не включайте password_hash в ответ

	http.Error(w, "Profile not implemented", http.StatusNotImplemented)
}

// HealthHandler проверяет состояние сервиса
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем подключение к БД
	if db != nil {
		if err := db.Ping(); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}
	}

	// Возвращаем статус OK
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "ok",
		"message": "Service is running",
	}
	json.NewEncoder(w).Encode(response)
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
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	// TODO: Добавьте дополнительные проверки
	// - Используйте ValidateEmail() и ValidatePassword() из auth.go
	// - Проверьте длину username (например, минимум 3 символа)
	// - Проверьте что username содержит только допустимые символы

	return nil
}

// validateLoginRequest валидирует данные входа
func validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
