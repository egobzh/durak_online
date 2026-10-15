# Dev Docs — «Дурак Онлайн» (Бэкенд)

Документация разработчика по запуску, конфигурации и разработке бэкенд-сервиса карточной игры «Дурак Онлайн».

---

## 1. Архитектура и структура проекта

Проект организован по принципам чистой слоистой архитектуры (Clean/Hexagonal Architecture):

```
code/
├── cmd/
│   └── server/
│       └── main.go                  # Точка входа в приложение, инициализация зависимостей
├── docs/                            # Сгенерированные Swagger-спецификации (swagger.json, docs.go)
├── internal/
│   ├── config/                      # Загрузка переменных окружения (порты, БД, JWT, TTL)
│   ├── delivery/
│   │   └── http/
│   │       ├── handler.go           # Контроллеры / HTTP-хэндлеры для Auth и Profile
│   │       ├── router.go            # Регистрация маршрутов Gin и Swagger UI
│   │       └── middleware/
│   │           └── auth.go          # Проверка Bearer JWT и версий токена (инвалидация сессий)
│   ├── domain/
│   │   └── models.go                # Модели сущностей (User, PasswordResetToken), DTO и интерфейсы
│   ├── notifier/
│   │   └── console.go               # Реализация ResetLinkSender для логирования ссылок в консоль
│   ├── repository/
│   │   └── postgres/
│   │       └── postgres.go          # Подключение к PostgreSQL, автосоздание схемы и репозитории
│   └── service/
│       └── auth.go                  # Бизнес-логика: регистрация, сброс пароля, профиль
├── pkg/
│   ├── hasher/
│   │   └── hasher.go                # Хэширование паролей bcrypt
│   └── token/
│       └── jwt.go                   # Генерация и валидация JWT
├── dev-docs.md                      # Настоящее руководство разработчика
├── docker-compose.yml               # Запуск сервиса и базы данных PostgreSQL
└── Dockerfile                       # Multi-stage сборка Go-бинарника в минимальном Alpine
```

---

## 2. Реализованный функционал (MVP Этап 1)

1. **Регистрация (`POST /api/v1/auth/register`)**:
   - Без авторизации;
   - Проверка уникальности e-mail и логина;
   - Хэширование пароля (`bcrypt`);
   - Начисление стартового капитала: **5000 кредитов** и **1200 рейтинга** (по SRS / PRD);
   - Автоматическая генерация JWT-токена на 24 часа.
2. **Запрос сброса пароля (`POST /api/v1/auth/forgot-password`)**:
   - Без авторизации;
   - Защита от перебора аккаунтов (*anti-enumeration*): возвращается одинаковый статус независимо от того, существует e-mail или нет;
   - Токен сброса с временем жизни (TTL) 30 минут;
   - Вывод ссылки и токена через абстракцию `ResetLinkSender` в консоль.
3. **Подтверждение сброса пароля (`POST /api/v1/auth/reset-password`)**:
   - Без авторизации;
   - Проверка срока действия и однократности использования токена;
   - **Инвалидация всех активных сессий**: инкрементируется `token_version` пользователя, после чего старые токены на других девайсах становятся невалидными (401).
4. **Профиль игрока (`GET /api/v1/profile/me`)**:
   - Доступен строго по Bearer токену (`Authorization: Bearer <token>`);
   - Возвращает: `id`, `email`, `username`, `credits`, `rating`, `total_matches`, `wins`.
5. **Интерактивная Swagger документация**:
   - UI доступен по адресу: `http://localhost:8080/swagger/index.html`.

---

## 3. Требования к окружению

Для работы с проектом потребуется:
- **Docker** и **Docker Compose** (рекомендуемый способ запуска);
- Или локально: **Go 1.22+** и установленный сервер **PostgreSQL 14+**.

---

## 4. Запуск проекта

### 4.1 Быстрый запуск через Docker Compose (Рекомендуется)

Перейдите в директорию `code`:
```bash
cd code
```

Запустите сборку и контейнеры в фоне:
```bash
docker compose up --build -d
```

Проверить статус контейнеров:
```bash
docker compose ps
```

Просмотр логов бэкенда (включая ссылки сброса пароля):
```bash
docker compose logs -f backend
```

Остановка контейнеров:
```bash
docker compose down
```

Если необходимо удалить сохраненные данные базы:
```bash
docker compose down -v
```

---

### 4.2 Локальный запуск без Docker

1. Убедитесь, что PostgreSQL запущен и создана БД (например `durak_db`).
2. Задайте переменные окружения (или используйте значения по умолчанию):
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/durak_db?sslmode=disable"
   export SERVER_PORT="8080"
   export JWT_SECRET="your-local-secret"
   ```
3. Запустите приложение:
   ```bash
   go run ./cmd/server/main.go
   ```

---

## 5. Генерация Swagger-документации

В коде хэндлеров и `main.go` расставлены декларативные аннотации `swaggo`.

### 5.1 Установка swag CLI (однократно)
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```
*Убедитесь, что путь `$(go env GOPATH)/bin` добавлен в переменную окружения `PATH`.*

### 5.2 Перегенерация документации
Выполните команду из корня папки `code`:
```bash
swag init -g ./cmd/server/main.go -o ./docs --parseDependency --parseInternal
```

После перегенерации в каталоге `code/docs/` обновятся файлы `docs.go`, `swagger.json`, `swagger.yaml`.
При перезапуске сервера интерактивный Swagger UI будет доступен по адресу:
👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

---

## 6. Примеры работы с API (cURL / HTTP)

### 6.1 Регистрация
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "player1@example.com",
    "username": "CardMaster",
    "password": "Password123"
  }'
```
**Ответ (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "c1f7b82e-9d22-4416-...",
    "email": "player1@example.com",
    "username": "CardMaster",
    "credits": 5000,
    "rating": 1200,
    "total_matches": 0,
    "wins": 0,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

### 6.2 Получение своего профиля
```bash
curl -X GET http://localhost:8080/api/v1/profile/me \
  -H "Authorization: Bearer <ВСТАВЬТЕ_ВАШ_ТОКЕН>"
```
**Ответ (200 OK):**
```json
{
  "id": "c1f7b82e-9d22-4416-...",
  "email": "player1@example.com",
  "username": "CardMaster",
  "credits": 5000,
  "rating": 1200,
  "total_matches": 0,
  "wins": 0
}
```

### 6.3 Запрос сброса пароля
```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "player1@example.com"
  }'
```
В логах контейнера (`docker compose logs backend`) появится сообщение:
```
=======================================================
[PASSWORD RESET NOTIFICATION]
To: player1@example.com
Reset Token: 4945d8b2-5f72-4638-a89e-04f71a415a77
Reset Link:  http://localhost:8080/api/v1/auth/reset-password?token=4945d8b2-5f72-4638-a89e-04f71a415a77
Expires in:  30 minutes
=======================================================
```

### 6.4 Подтверждение смены пароля
```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "4945d8b2-5f72-4638-a89e-04f71a415a77",
    "new_password": "NewSecretPassword123"
  }'
```
*После успешного сброса старый токен авторизации перестает действовать.*

---

## 7. Конфигурация через переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `SERVER_PORT` | `8080` | Порт HTTP сервера |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/durak_db?sslmode=disable` | Строка подключения к PostgreSQL |
| `JWT_SECRET` | `super-secret-durak-jwt-key-2026` | Секретный ключ подписи JWT |
| `JWT_EXPIRATION_HOURS` | `24` | Срок жизни JWT токена в часах |
| `RESET_BASE_URL` | `http://localhost:8080/api/v1/auth/reset-password` | Базовый URL для генерации ссылки восстановления |
| `RESET_TOKEN_TTL_MINUTES` | `30` | Срок действия токена сброса пароля в минутах |

---

## 8. Архитектурные абстракции: интерфейс отправщика ссылок

В пакете `internal/domain/models.go` объявлен интерфейс:
```go
type ResetLinkSender interface {
    SendPasswordResetLink(ctx context.Context, email, token string) error
}
```
На этапе MVP используется `ConsoleResetLinkSender` (`internal/notifier/console.go`), который печатает ссылку в stdout.
Для подключения реальной почты (SMTP, Mailgun, SendGrid) достаточно создать структуру, реализующую данный интерфейс, и передать её в `NewAuthService` в `cmd/server/main.go` без изменений бизнес-логики.
