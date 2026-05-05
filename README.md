# rst — Restaurant Table Service API

REST API do zarządzania restauracją: menu, stoliki, zamówienia, wywołania kelnera i powiadomienia w czasie rzeczywistym przez WebSocket.

## Stack

- **Go 1.26** — Gin, pgx, golang-migrate
- **PostgreSQL** — baza danych (migracje embedded)
- **Redis** — cache CORS, kolejka zadań asynchronicznych (Asynq)
- **WebSocket** — powiadomienia real-time dla właściciela restauracji
- **Gmail SMTP** — wysyłka maili (reset hasła)

## Wymagania

- Go 1.26+
- PostgreSQL
- Redis

## Uruchomienie

```bash
cp .env.example .env   # uzupełnij zmienne
go run ./cmd/api
```

Serwer przy starcie automatycznie uruchamia migracje i worker kolejki.

## Zmienne środowiskowe

| Zmienna | Domyślna | Opis |
|---------|----------|------|
| `PORT` | `8080` | Port serwera |
| `DATABASE_URL` | — | Connection string PostgreSQL |
| `ENV` | `development` | Tryb (`development` / `production`) |
| `JWT_SECRET` | `change-me-in-production` | Sekret podpisywania JWT |
| `FRONTEND_URL` | `http://localhost:3000` | Adres frontendu (CORS + linki w mailach) |
| `REDIS_ADDR` | `localhost:6379` | Adres Redis |
| `GMAIL_FROM` | — | Adres Gmail nadawcy |
| `GMAIL_APP_PASSWORD` | — | Hasło aplikacji Gmail (nie logowania) |

## API

### Autentykacja

| Metoda | Ścieżka | Opis |
|--------|---------|------|
| `POST` | `/api/auth/login` | Logowanie |
| `POST` | `/api/auth/logout` | Wylogowanie |
| `POST` | `/api/auth/refresh` | Odświeżenie access tokena |
| `POST` | `/api/auth/forgot-password` | Żądanie resetu hasła |
| `POST` | `/api/auth/reset-password` | Ustawienie nowego hasła tokenem |

### Użytkownicy

| Metoda | Ścieżka | Opis |
|--------|---------|------|
| `POST` | `/api/users` | Rejestracja |

### Restauracje `🔒`

| Metoda | Ścieżka | Opis |
|--------|---------|------|
| `GET` | `/api/restaurants/:id` | Pobierz restaurację |
| `POST` | `/api/restaurants` | Utwórz restaurację |
| `GET` | `/api/restaurants/my` | Restauracje zalogowanego użytkownika |

### Menu `🔒`

| Metoda | Ścieżka | Opis |
|--------|---------|------|
| `GET` | `/api/restaurants/:id/menus` | Lista menu |
| `GET` | `/api/restaurants/:id/menu-items` | Wszystkie pozycje menu restauracji |
| `POST` | `/api/restaurants/:id/menus` | Utwórz menu |
| `POST` | `/api/menus/:menuId/items` | Dodaj pozycję do menu |
| `DELETE` | `/api/menu-items/:itemId` | Usuń pozycję |

### Stoliki `🔒`

| Metoda | Ścieżka | Opis |
|--------|---------|------|
| `GET` | `/api/restaurants/:id/tables` | Lista stolików |
| `POST` | `/api/restaurants/:id/tables` | Utwórz stolik |
| `PATCH` | `/api/restaurants/:id/tables/:tableId` | Aktualizuj stolik |
| `DELETE` | `/api/restaurants/:id/tables/:tableId` | Usuń stolik |

### Zamówienia

| Metoda | Ścieżka | Auth | Opis |
|--------|---------|------|------|
| `POST` | `/api/restaurants/:id/tables/:tableId/orders` | — | Złóż zamówienie |
| `GET` | `/api/restaurants/:id/tables/:tableId/orders` | — | Zamówienia stolika |
| `GET` | `/api/restaurants/:id/orders` | 🔒 | Wszystkie zamówienia restauracji |
| `PATCH` | `/api/restaurants/:id/orders/:orderId/status` | 🔒 | Zmień status zamówienia |

### Wywołania kelnera

| Metoda | Ścieżka | Auth | Opis |
|--------|---------|------|------|
| `POST` | `/api/restaurants/:id/tables/:tableId/calls` | — | Wywołaj kelnera |
| `GET` | `/api/restaurants/:id/calls` | 🔒 | Lista wywołań |
| `PATCH` | `/api/restaurants/:id/calls/:callId/status` | 🔒 | Zmień status wywołania |

### WebSocket

| Ścieżka | Auth | Opis |
|---------|------|------|
| `GET /api/ws/restaurants/:id` | JWT w `?token=` | Właściciel — odbiera zdarzenia real-time |
| `GET /api/ws/restaurants/:id/tables/:tableId` | — | Stolik — odbiera zdarzenia real-time |

## Testy

```bash
# Jednostkowe
go test ./internal/modules/...

# Integracyjne (wymaga Dockera)
go test -tags integration ./internal/modules/...
```

Testy jednostkowe korzystają z `go-sqlmock`, integracyjne z `testcontainers-go` (Postgres w kontenerze).

## Struktura projektu

```
cmd/api/          # Punkt wejścia
internal/
  database/       # Połączenie DB, migracje (embedded SQL)
  modules/
    auth/         # JWT, sesje, reset hasła
    user/         # Rejestracja
    restaurant/   # Restauracje + sub-moduły: menu, table, order, call
  ws/             # Handler HTTP WebSocket
  testhelper/     # Pomocniki do testów integracyjnych
pkg/
  config/         # Ładowanie zmiennych środowiskowych
  logger/         # Strukturalne logi (slog)
  mailer/         # Gmail SMTP
  middleware/     # Auth JWT, dynamiczny CORS (cache Redis)
  worker/         # Asynq — serwer + procesor zadań email
  ws/             # Hub WebSocket, klienci, zdarzenia
```
