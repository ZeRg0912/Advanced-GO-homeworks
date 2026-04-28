package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	// CGO реализация, либо надо подключать CGO, либо использовать чистую реализацию на Go, например modernc.org/sqlite
	//_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

type Order struct {
	ID       int
	Customer string
	Products string
	Total    float64
	Status   string
}

type Message struct {
	Recipient string
	Text      string
}

// RepositoryWriter отвечает только за запись заказов.
type RepositoryWriter interface {
	Save(order Order) error
}

// RepositoryInitializer отвечает только за инициализацию хранилища.
type RepositoryInitializer interface {
	Init() error
}

// Notifier отвечает только за отправку уведомлений.
type Notifier interface {
	Send(message Message) error
}

type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) Init() error {
	_, err := r.db.Exec(`
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer TEXT NOT NULL,
		products TEXT NOT NULL,
		total REAL NOT NULL,
		status TEXT NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("init sqlite repository: %w", err)
	}

	return nil
}

func (r *SQLiteRepo) Save(order Order) error {
	_, err := r.db.Exec(
		"INSERT INTO orders (customer, products, total, status) VALUES (?, ?, ?, ?)",
		order.Customer,
		order.Products,
		order.Total,
		order.Status,
	)
	if err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	return nil
}

type EmailSender struct{}

func (s EmailSender) Send(message Message) error {
	fmt.Printf("Email отправлен клиенту %s: %s\n", message.Recipient, message.Text)
	return nil
}

type SMSSender struct{}

func (s SMSSender) Send(message Message) error {
	fmt.Printf("SMS отправлено клиенту %s: %s\n", message.Recipient, message.Text)
	return nil
}

type OrderService struct {
	repository RepositoryWriter
	notifier   Notifier
}

func NewOrderService(repository RepositoryWriter, notifier Notifier) *OrderService {
	return &OrderService{
		repository: repository,
		notifier:   notifier,
	}
}

func (s *OrderService) CreateOrder(customer string, products []string, total float64) error {
	order := Order{
		Customer: customer,
		Products: strings.Join(products, ", "),
		Total:    total,
		Status:   "pending",
	}

	if err := s.repository.Save(order); err != nil {
		return err
	}

	message := Message{
		Recipient: customer,
		Text:      fmt.Sprintf("Ваш заказ на сумму %.2f создан и ожидает обработки", total),
	}

	if err := s.notifier.Send(message); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}

	return nil
}

func main() {
	//db, err := sql.Open("sqlite3", "orders.db")
	db, err := sql.Open("sqlite", "orders.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repository := NewSQLiteRepo(db)

	if err := repository.Init(); err != nil {
		log.Fatal(err)
	}

	emailService := NewOrderService(repository, EmailSender{})
	if err := emailService.CreateOrder("Иван", []string{"apple", "banana"}, 10.5); err != nil {
		log.Fatal(err)
	}

	smsService := NewOrderService(repository, SMSSender{})
	if err := smsService.CreateOrder("Мария", []string{"orange", "mango"}, 25.75); err != nil {
		log.Fatal(err)
	}
}
