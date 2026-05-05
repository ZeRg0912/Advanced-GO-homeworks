package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Глобальная переменная для подключения к БД
var db *sql.DB

// InitDB инициализирует подключение к базе данных
func InitDB() error {
	// TODO: Реализуйте подключение к PostgreSQL
	//
	// Что нужно сделать:
	// 1. Составьте строку подключения используя fmt.Sprintf()
	//    Формат: "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"
	// 2. Получите параметры из переменных окружения с помощью getEnv()
	// 3. Откройте соединение с sql.Open("postgres", connStr)
	// 4. Проверьте подключение с помощью db.Ping()
	// 5. Обработайте ошибки на каждом шаге
	//
	// Переменные окружения: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "secure_service"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	return nil
}

// CloseDB закрывает соединение с базой данных
func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// CreateUser создает нового пользователя в базе данных
func CreateUser(email, username, passwordHash string) (*User, error) {
	// TODO: Реализуйте создание пользователя
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос для защиты от SQL-инъекций!
	//
	// Что нужно сделать:
	// 1. Создайте SQL запрос с плейсхолдерами $1, $2, $3
	//    INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at
	// 2. Выполните запрос с db.QueryRow(query, email, username, passwordHash)
	// 3. Считайте результат в переменные user.ID и user.CreatedAt
	// 4. Заполните остальные поля структуры User
	// 5. Обработайте ошибки
	//
	// НИКОГДА не используйте fmt.Sprintf для построения SQL запросов!

	return nil, fmt.Errorf("not implemented - реализуйте создание пользователя")
}

// GetUserByEmail находит пользователя по email
func GetUserByEmail(email string) (*User, error) {
	// TODO: Реализуйте поиск пользователя по email
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос!
	//
	// Что нужно сделать:
	// 1. Создайте SQL запрос с плейсхолдером $1
	//    SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1
	// 2. Выполните запрос с db.QueryRow(query, email)
	// 3. Считайте все поля в структуру User с помощью Scan()
	// 4. Обработайте случай sql.ErrNoRows (пользователь не найден)
	//
	// Подсказка: используйте sql.ErrNoRows для проверки отсутствия результата

	return nil, fmt.Errorf("not implemented - реализуйте поиск пользователя по email")
}

// GetUserByID находит пользователя по ID
func GetUserByID(userID int) (*User, error) {
	// TODO: Реализуйте поиск пользователя по ID
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос!
	//
	// Что нужно сделать:
	// 1. Создайте SQL запрос для поиска по ID
	// 2. НЕ включайте password_hash в SELECT (он не нужен для профиля)
	// 3. Выполните запрос и обработайте результат
	//
	// Запрос: SELECT id, email, username, created_at FROM users WHERE id = $1

	return nil, fmt.Errorf("not implemented - реализуйте поиск пользователя по ID")
}

// UserExistsByEmail проверяет, существует ли пользователь с данным email
func UserExistsByEmail(email string) (bool, error) {
	// TODO: Реализуйте проверку существования пользователя
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос!
	//
	// Что нужно сделать:
	// 1. Используйте SQL функцию EXISTS для эффективной проверки
	//    SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	// 2. Результат будет булевым значением
	// 3. Считайте результат в переменную типа bool
	//
	// Это эффективнее чем получать полную запись пользователя

	return false, fmt.Errorf("not implemented - реализуйте проверку существования пользователя")
}

// GetDB возвращает подключение к базе данных (для тестирования)
func GetDB() *sql.DB {
	return db
}
