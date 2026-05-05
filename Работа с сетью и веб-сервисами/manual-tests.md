# Ручное тестирование tasks-api

Перед проверкой запустить сервер:

```bash
go run ./cmd/server
```

Команды ниже рассчитаны на терминал Visual Studio Code / PowerShell, поэтому используется `curl.exe`. Для запросов с JSON-телом используется `curl.exe --%`: это нужно, чтобы PowerShell не ломал кавычки внутри JSON. В Linux/macOS можно заменить `curl.exe` на `curl` и использовать bash-вариант с одинарными кавычками.

Ответы с `id` и `created_at` примерные: `id` зависит от порядка создания задач, а `created_at` будет текущим временем сервера. Для сценариев с `/tasks/1` сначала выполните корректный `POST /tasks`.

## GET /health

Корректный сценарий:

```powershell
curl.exe -i http://localhost:8080/health
```

Ожидаемо:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{"status":"ok"}
```

Ошибочный сценарий, неподдерживаемый метод:

```powershell
curl.exe -i -X POST http://localhost:8080/health
```

Ожидаемо:

```http
HTTP/1.1 405 Method Not Allowed
Content-Type: application/json
```

```json
{"error":"method not allowed"}
```

## GET /tasks

Корректный сценарий:

```powershell
curl.exe -i http://localhost:8080/tasks
```

Ожидаемо для нового запуска сервера:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
[]
```

Ошибочный сценарий, неподдерживаемый метод:

```powershell
curl.exe -i -X PATCH http://localhost:8080/tasks
```

Ожидаемо:

```http
HTTP/1.1 405 Method Not Allowed
Content-Type: application/json
```

```json
{"error":"method not allowed"}
```

## POST /tasks

Корректный сценарий:

```powershell
curl.exe --% -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d "{\"title\":\"Купить продукты\",\"done\":false}"
```

Ожидаемо:

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

Ошибочный сценарий, пустой `title`:

```powershell
curl.exe --% -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d "{\"title\":\"\",\"done\":false}"
```

Ожидаемо:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json
```

```json
{"error":"title is required"}
```

## GET /tasks/{id}

Корректный сценарий:

```powershell
curl.exe -i http://localhost:8080/tasks/1
```

Ожидаемо:

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

Ошибочный сценарий, несуществующая задача:

```powershell
curl.exe -i http://localhost:8080/tasks/999
```

Ожидаемо:

```http
HTTP/1.1 404 Not Found
Content-Type: application/json
```

```json
{"error":"task not found"}
```

## PUT /tasks/{id}

Корректный сценарий:

```powershell
curl.exe --% -i -X PUT http://localhost:8080/tasks/1 -H "Content-Type: application/json" -d "{\"title\":\"Купить продукты и воду\",\"done\":true}"
```

Ожидаемо:

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

Ошибочный сценарий, некорректный `id`:

```powershell
curl.exe --% -i -X PUT http://localhost:8080/tasks/abc -H "Content-Type: application/json" -d "{\"title\":\"Текст\",\"done\":true}"
```

Ожидаемо:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json
```

```json
{"error":"task id must be a positive integer"}
```

## DELETE /tasks/{id}

Корректный сценарий:

```powershell
curl.exe -i -X DELETE http://localhost:8080/tasks/1
```

Ожидаемо:

```http
HTTP/1.1 204 No Content
Content-Type: application/json
```

Тело ответа отсутствует.

Ошибочный сценарий, повторное удаление той же задачи:

```powershell
curl.exe -i -X DELETE http://localhost:8080/tasks/1
```

Ожидаемо:

```http
HTTP/1.1 404 Not Found
Content-Type: application/json
```

```json
{"error":"task not found"}
```

## GET /unknown

Ошибочный сценарий, неизвестный маршрут:

```powershell
curl.exe -i http://localhost:8080/unknown
```

Ожидаемо:

```http
HTTP/1.1 404 Not Found
Content-Type: application/json
```

```json
{"error":"resource not found"}
```
