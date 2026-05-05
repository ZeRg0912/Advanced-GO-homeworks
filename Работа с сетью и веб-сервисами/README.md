# tasks-api

REST API на Go для управления списком задач. Сервис работает в памяти, без БД, принимает и отдаёт только JSON.

## Требования

- Go 1.20+; рекомендуется Go 1.21+
- curl, curl.exe или Postman для ручной проверки

## Запуск

```bash
go run ./cmd/server
```

Сервер запускается на `http://localhost:8080`.

Быстрая проверка в терминале Visual Studio Code / PowerShell:

```powershell
curl.exe -i http://localhost:8080/health
```

Ожидаемый ответ:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{"status":"ok"}
```

## Модель данных

```json
{
  "id": 1,
  "title": "Купить продукты",
  "done": false,
  "created_at": "2026-05-04T17:00:00Z"
}
```

Поля:

| Поле | Тип | Описание |
| --- | --- | --- |
| `id` | number | Генерируется на сервере. |
| `title` | string | Обязательное поле. Пустая строка не допускается. |
| `done` | boolean | Статус выполнения задачи. |
| `created_at` | string | UTC-время создания в формате RFC3339/ISO8601. Генерируется на сервере. |

`id` и `created_at` в примерах являются примерными: фактический `id` зависит от порядка создания задач, а `created_at` выставляется по текущему времени сервера.

## Формат ошибок

Все ошибки возвращаются в едином JSON-формате:

```json
{"error":"сообщение"}
```

Для всех ответов выставляется заголовок:

```http
Content-Type: application/json
```

## Эндпоинты

| Метод | Путь | Назначение | Успешный код |
| --- | --- | --- | --- |
| `GET` | `/health` | Health-check сервиса. | `200` |
| `GET` | `/tasks` | Получить список задач. | `200` |
| `POST` | `/tasks` | Создать задачу. | `201` |
| `GET` | `/tasks/{id}` | Получить задачу по идентификатору. | `200` |
| `PUT` | `/tasks/{id}` | Обновить задачу целиком. | `200` |
| `DELETE` | `/tasks/{id}` | Удалить задачу. | `204` |

Коды ошибок:

| Код | Когда используется |
| --- | --- |
| `400` | Некорректный JSON, пустое обязательное поле `title`, некорректный `id`. |
| `404` | Задача или ресурс не найдены. |
| `405` | Метод не поддерживается для маршрута. |
| `500` | Внутренняя ошибка сервера. |

## Примеры curl.exe для Windows / PowerShell

### GET /health

```powershell
curl.exe -i http://localhost:8080/health
```

Ответ:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{"status":"ok"}
```

### GET /tasks

```powershell
curl.exe -i http://localhost:8080/tasks
```

Ответ для пустого списка:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
[]
```

### POST /tasks

```powershell
curl.exe --% -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d "{\"title\":\"Купить продукты\",\"done\":false}"
```

Ответ:

```http
HTTP/1.1 201 Created
Content-Type: application/json
```

```json
{
  "id": 1,
  "title": "Купить продукты",
  "done": false,
  "created_at": "2026-05-04T17:00:00Z"
}
```

### GET /tasks/{id}

```powershell
curl.exe -i http://localhost:8080/tasks/1
```

Ответ:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "id": 1,
  "title": "Купить продукты",
  "done": false,
  "created_at": "2026-05-04T17:00:00Z"
}
```

### PUT /tasks/{id}

```powershell
curl.exe --% -i -X PUT http://localhost:8080/tasks/1 -H "Content-Type: application/json" -d "{\"title\":\"Купить продукты и воду\",\"done\":true}"
```

Ответ:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "id": 1,
  "title": "Купить продукты и воду",
  "done": true,
  "created_at": "2026-05-04T17:00:00Z"
}
```

### DELETE /tasks/{id}

```powershell
curl.exe -i -X DELETE http://localhost:8080/tasks/1
```

Ответ:

```http
HTTP/1.1 204 No Content
Content-Type: application/json
```

Тело ответа отсутствует.

## Ручное тестирование

Полный набор команд для корректных и ошибочных сценариев находится в [`manual-tests.md`](manual-tests.md).

## OpenAPI

Описание API находится в [`openapi.yaml`](openapi.yaml).

## Структура проекта

```text
tasks-api/
  cmd/server/main.go
  internal/handlers/tasks.go
  internal/models/task.go
  internal/storage/storage.go
  internal/storage/memory/memory.go
  internal/http/middleware.go
  internal/handlers/tasks_test.go
  manual-tests.md
  openapi.yaml
  README.md
  go.mod
```

## Проверка сборки и тестов

```bash
go test ./...
```
