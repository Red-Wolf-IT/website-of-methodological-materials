# Website of Methodological Materials

Backend API на Go для каталога методических материалов.

**Архитектура:** Handlers → Services → Repositories → PostgreSQL  
**Асинхронность:** worker для счётчика просмотров через канал `viewsChan`

## Стек

| Компонент | Технология |
|---|---|
| Язык | Go 1.25+ |
| HTTP-роутер | [chi](https://github.com/go-chi/chi) |
| БД | PostgreSQL + [pgx](https://github.com/jackc/pgx) |
| Валидация | [go-playground/validator](https://github.com/go-playground/validator) |
| Конфиг | env + [godotenv](https://github.com/joho/godotenv) |

## Структура проекта

```
cmd/app/                    — точка входа, graceful shutdown
internal/
  config/                   — переменные окружения
  db/                       — пул соединений PostgreSQL
  server/                   — chi-роутер и маршруты
  handlers/                 — HTTP-контроллеры
  service/                  — бизнес-логика
  repository/postgres/      — SQL-запросы
  models/                   — сущности и DTO
  middleware/               — логирование, recover, admin-auth
  validator/                — обёртка над validator
  worker/                   — асинхронный счётчик просмотров
  storage/                  — сохранение файлов на диск
storage/
  migrations/               — SQL-миграции
  seeds/                    — тестовые данные
  uploads/                  — загруженные вложения
scripts/
  smoke_test.ps1            — быстрая проверка API
web/                        — Vue 3 фронтенд (каталог для пользователей)
```

## Быстрый старт

### 1. PostgreSQL

Создайте базу данных:

```bash
psql -U postgres -c "CREATE DATABASE mrepo;"
```

Примените миграции по порядку:

```bash
psql -U postgres -d mrepo -f storage/migrations/001_create_manuals.sql
psql -U postgres -d mrepo -f storage/migrations/002_create_tags.sql
psql -U postgres -d mrepo -f storage/migrations/003_create_manual_tags.sql
```

Загрузите тестовые данные:

```bash
psql -U postgres -d mrepo -f storage/seeds/001_sample_data.sql
```

### 2. Конфигурация

```bash
cp .env.example .env
```

Заполните `.env`:

| Переменная | Описание | Пример |
|---|---|---|
| `SERVER_ADDR` | Адрес HTTP-сервера | `:8080` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `postgres` |
| `DB_PASSWORD` | Пароль БД | `pass` |
| `DB_NAME` | Имя базы | `mrepo` |
| `DB_SSLMODE` | SSL-режим pgx | `disable` |
| `ADMIN_TOKEN` | Токен для админ-эндпоинтов | `dev-admin-secret` |
| `STORAGE_DIR` | Папка для вложений | `storage/uploads` |

### 3. Запуск

```bash
go run ./cmd/app
```

Проверка:

```bash
curl http://localhost:8080/health
# {"status":"ok","database":"ok"}
```

Быстрый smoke-тест всех основных сценариев:

```powershell
powershell -File scripts/smoke_test.ps1
```

### 4. Веб-интерфейс (для пользователей)

В папке `web/` — Vue 3 + Vite + Tailwind. Каталог, поиск, просмотр материалов, добавление новых.

**Терминал 1 — backend:**

```bash
go run ./cmd/app
```

**Терминал 2 — frontend:**

```bash
cd web
npm install
npm run dev
```

Откройте в браузере: **http://localhost:5173**

Запросы проксируются на API через `/api` → `localhost:8080`.

**Возможности интерфейса:**

| Страница | URL | Что делает |
|---|---|---|
| Каталог | `/` | Список, поиск, фильтр по тегу, сортировка, пагинация |
| Материал | `/manuals/{id}` | Полный текст, теги, скачивание вложения |
| Добавить | `/create` | Форма создания + выбор/создание тегов |

Сборка для продакшена:

```bash
cd web && npm run build
# статика в web/dist/
```

---

## Схема БД

```
manuals (1) ──< manual_tags >── (1) tags
```

| Таблица | Описание |
|---|---|
| `manuals` | Материалы: title, author, content, file_path, views_count |
| `tags` | Справочник тегов (уникальное `name`) |
| `manual_tags` | Связь M:N, FK с `ON DELETE CASCADE` |

---

## Формат ответов

**Успех:**

```json
{"data": { ... }}
```

**Ошибка:**

```json
{"error": {"message": "..."}}
```

**Валидация (400):**

```json
{
  "error": {
    "message": "validation failed",
    "fields": [
      {"field": "title", "message": "is required"}
    ]
  }
}
```

---

## API

### Сводная таблица

| Метод | Путь | Auth | Описание |
|---|---|---|---|
| GET | `/health` | — | Проверка сервера и БД |
| GET | `/tags` | — | Список тегов |
| POST | `/tags` | — | Создать тег |
| POST | `/manuals` | — | Создать материал |
| GET | `/manuals` | — | Список с фильтрами |
| GET | `/manuals/{id}` | — | Получить материал (+ теги, +1 просмотр) |
| POST | `/manuals/{id}/tags` | — | Привязать теги |
| GET | `/uploads/{filename}` | — | Скачать вложение |
| PUT | `/manuals/{id}` | Admin | Обновить материал |
| DELETE | `/manuals/{id}` | Admin | Удалить материал |
| POST | `/manuals/{id}/attachment` | Admin | Загрузить файл |

> **Admin** — заголовок `X-Admin-Token: <ADMIN_TOKEN>`

---

### Healthcheck

```bash
curl http://localhost:8080/health
```

---

### Теги

**Список тегов**

```bash
curl http://localhost:8080/tags
```

**Создать тег**

```bash
curl -X POST http://localhost:8080/tags \
  -H "Content-Type: application/json" \
  -d '{"name":"microservices"}'
```

Ответ `201`: `{"data":{"id":6,"name":"microservices"}}`  
Дубликат имени → `409`

---

### Материалы

**Создать**

```bash
curl -X POST http://localhost:8080/manuals \
  -H "Content-Type: application/json" \
  -d '{"title":"Название","author":"Автор","content":"Текст методички"}'
```

**Список с фильтрами**

```bash
# все (page=1, limit=20 по умолчанию)
curl "http://localhost:8080/manuals"

# фильтр по тегу + сортировка по популярности
curl "http://localhost:8080/manuals?tag_id=1&sort=popular"

# поиск по автору
curl "http://localhost:8080/manuals?author=Иванов"

# полнотекстовый поиск (title, content, author)
curl "http://localhost:8080/manuals?q=PostgreSQL"

# пагинация
curl "http://localhost:8080/manuals?page=1&limit=2"
```

Ответ:

```json
{
  "data": {
    "items": [ ... ],
    "total": 4,
    "page": 1,
    "limit": 20
  }
}
```

**Получить по ID** (асинхронно увеличивает `views_count`)

```bash
curl http://localhost:8080/manuals/a1000000-0000-4000-8000-000000000001
```

Ответ включает массив `tags`.

**Привязать теги**

```bash
curl -X POST http://localhost:8080/manuals/{id}/tags \
  -H "Content-Type: application/json" \
  -d '{"tag_ids":[1,2,3]}'
```

---

### Админ-операции

Все запросы ниже требуют заголовок:

```
X-Admin-Token: dev-admin-secret
```

**Обновить материал**

```bash
curl -X PUT http://localhost:8080/manuals/{id} \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: dev-admin-secret" \
  -d '{"title":"Новое название","author":"Автор","content":"Новый текст"}'
```

> Поле `file_path` опционально — если не передано, существующее вложение сохраняется.

**Удалить материал**

```bash
curl -X DELETE http://localhost:8080/manuals/{id} \
  -H "X-Admin-Token: dev-admin-secret"
```

Ответ: `204 No Content`. Связи в `manual_tags` удаляются каскадом, файл — с диска.

**Загрузить вложение** (multipart, поле `file`, до 10 MB)

```bash
curl -X POST http://localhost:8080/manuals/{id}/attachment \
  -H "X-Admin-Token: dev-admin-secret" \
  -F "file=@document.pdf"
```

**Скачать вложение**

```bash
curl http://localhost:8080/uploads/{filename}
```

---

## Postman

1. Создайте Environment с переменными:
   - `base_url` = `http://localhost:8080`
   - `admin_token` = `dev-admin-secret`
   - `manual_id` = UUID материала из seed

2. Для админ-запросов добавьте header:
   - Key: `X-Admin-Token`
   - Value: `{{admin_token}}`

3. Для загрузки файла: Body → form-data → key `file` (тип File).

---

## Сценарий демонстрации (~5 мин)

Рекомендуемый порядок для защиты/презентации:

```
1. GET  /health                          → сервер и БД работают
2. GET  /manuals?tag_id=1&sort=popular   → список + фильтр + сортировка
3. GET  /manuals/{seed-id}               → детали с тегами
4. GET  /manuals/{seed-id}  (повторно)   → views_count вырос (worker)
5. POST /tags                            → создать новый тег
6. POST /manuals                         → создать материал
7. POST /manuals/{id}/tags               → привязать теги
8. PUT  /manuals/{id}  (+ Admin-Token)   → обновить
9. POST /manuals/{id}/attachment         → загрузить PDF/txt
10. GET /uploads/{filename}              → скачать файл
11. DELETE /manuals/{id} (+ Admin-Token) → удалить (без токена → 401)
```

Seed ID для шага 3: `a1000000-0000-4000-8000-000000000001` («Введение в Go», теги: go, backend, tutorial).

---

## Архитектурные решения

| Задача | Решение |
|---|---|
| M:N теги | Junction-таблица `manual_tags` |
| Счётчик просмотров | Горутина-worker + канал, неблокирующая отправка из handler |
| Динамические фильтры | SQL с плейсхолдерами `$1, $2...` без конкатенации значений |
| Админ-доступ | Middleware `X-Admin-Token` на группе роутов |
| Файлы | Локальный диск `storage/uploads/`, защита от path traversal |
| Остановка сервера | `Shutdown` → отмена context worker → drain канала |

---

## Разработка

```bash
# сборка
go build ./...

# запуск
go run ./cmd/app

# smoke-тест (сервер должен быть запущен)
powershell -File scripts/smoke_test.ps1
```
