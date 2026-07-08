# URL Shortener

Production-ready сервис для сокращения ссылок на Go.

Поддерживает регистрацию пользователей, JWT-авторизацию, создание коротких URL (с custom alias или автогенерацией), redirect и cache-aside через Redis.

## Стек

| Компонент | Технология |
|-----------|------------|
| HTTP | [Echo](https://echo.labstack.com/) |
| БД | PostgreSQL 16 |
| Кэш | Redis 7 |
| Auth | JWT (HS256) + bcrypt |
| Логирование | zap |
| Конфигурация | viper |
| Миграции | golang-migrate (SQL) |
| Инфра | Docker + docker-compose |

## Быстрый старт (Docker)

```bash
docker compose -f deployments/docker-compose.yml up --build -d
```

Сервис будет доступен на `http://localhost:8080`.

Проверка:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

Остановка:

```bash
docker compose -f deployments/docker-compose.yml down
```

## Локальный запуск

**Требования:** Go 1.23+, PostgreSQL, Redis.

```bash
cp configs/config.example.yaml configs/config.yaml
go mod tidy
go run ./cmd/api
```

Миграции применяются автоматически при старте (`migrations.run_on_startup: true`).

Ручной запуск миграций:

```bash
make migrate-up
make migrate-down
```

## Makefile

| Команда | Описание |
|---------|----------|
| `make docker-up` | Поднять все сервисы в Docker |
| `make docker-down` | Остановить контейнеры |
| `make docker-logs` | Логи API |
| `make run` | Запуск локально |
| `make build` | Сборка бинарника |
| `make test` | Тесты |
| `make migrate-up` | Применить миграции |
| `make migrate-down` | Откатить миграции |

## Конфигурация

Пример — `configs/config.example.yaml`. Локальный конфиг `configs/config.yaml` не коммитится (см. `.gitignore`).

Переменные окружения с префиксом `URLSHORTENER_` перекрывают значения из YAML:

```bash
URLSHORTENER_SERVER_PORT=9090
URLSHORTENER_JWT_SECRET=your-super-secret-key-at-least-32-chars
URLSHORTENER_APP_BASE_URL=http://localhost:8080
```

Основные секции:

| Секция | Назначение |
|--------|------------|
| `server` | host, port, таймауты |
| `app.base_url` | базовый URL для `short_url` в ответах |
| `postgres` | подключение к PostgreSQL |
| `redis` | подключение к Redis |
| `cache.url_ttl` | TTL кэша redirect в Redis |
| `jwt` | secret и время жизни токена |
| `migrations` | путь к SQL-файлам, автозапуск |

## API

Базовый URL: `http://localhost:8080`

### Health

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| GET | `/health` | нет | Liveness probe |
| GET | `/ready` | нет | Readiness (PG + Redis) |

### Auth

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| POST | `/api/v1/auth/register` | нет | Регистрация |
| POST | `/api/v1/auth/login` | нет | Вход |
| GET | `/api/v1/me` | JWT | Текущий пользователь |

**Register / Login** — тело запроса:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Ответ:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### URLs

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| POST | `/api/v1/urls` | JWT | Создать короткую ссылку |
| GET | `/api/v1/urls` | JWT | Список своих ссылок |
| DELETE | `/api/v1/urls/:alias` | JWT | Удалить ссылку |
| GET | `/:alias` | нет | Redirect (302) |

**Create** — тело запроса:

```json
{
  "original_url": "https://google.com",
  "alias": "goog",
  "expires_at": "2026-12-31T00:00:00Z"
}
```

Поля `alias` и `expires_at` опциональны. Если `alias` не указан — генерируется автоматически (8 символов).

Ответ:

```json
{
  "alias": "goog",
  "original_url": "https://google.com",
  "short_url": "http://localhost:8080/goog",
  "created_at": "2026-07-08T08:06:12Z"
}
```

### Примеры (curl)

```bash
# регистрация
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# login → сохранить token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' | jq -r .token)

# создать ссылку
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://google.com","alias":"goog"}'

# redirect
curl -s -o /dev/null -w "%{http_code} %{redirect_url}\n" http://localhost:8080/goog

# список ссылок
curl http://localhost:8080/api/v1/urls -H "Authorization: Bearer $TOKEN"

# удалить
curl -X DELETE http://localhost:8080/api/v1/urls/goog -H "Authorization: Bearer $TOKEN"
```

> **Windows PowerShell:** используй `curl.exe` и экранирование JSON: `'{\"email\":\"...\"}'`

## Архитектура

```
cmd/api/          — точка входа
internal/
  app/            — wiring, graceful shutdown
  config/         — viper
  domain/         — сущности и доменные ошибки
  service/        — бизнес-логика
  repository/     — PostgreSQL
  cache/          — Redis
  server/         — Echo, handlers, middleware
  migrate/        — обёртка golang-migrate
migrations/       — SQL-миграции
deployments/      — Dockerfile, docker-compose
```

Поток redirect (cache-aside):

```
GET /:alias → Redis → (miss) → PostgreSQL → Redis SET → 302 Redirect
```

## Структура БД

**users** — id, email (unique), password_hash, created_at

**urls** — id, alias (unique), original_url, user_id, expires_at, created_at

## Ограничения alias

- длина: 3–32 символа
- символы: `a-z`, `A-Z`, `0-9`, `_`, `-`
- зарезервированы: `health`, `ready`, `api`

## Доработки

- rate limiting
- метрики (Prometheus)
- тесты (unit + integration)
- аналитика кликов
- CI/CD