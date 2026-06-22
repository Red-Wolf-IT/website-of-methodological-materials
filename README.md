# Website of Methodological Materials

Backend API на Go с архитектурой **Handlers → Services → Repositories**.

## Стек

- Go
- [chi](https://github.com/go-chi/chi) — HTTP-роутер
- PostgreSQL — [pgx](https://github.com/jackc/pgx)

## Структура

```
cmd/app/              — точка входа
internal/
  config/             — конфигурация из env
  server/             — chi-роутер
  handlers/           — HTTP-контроллеры
  service/            — бизнес-логика
  repository/postgres/  — работа с БД
  models/             — DTO/Entity
  db/                 — подключение к PostgreSQL
storage/migrations/   — SQL-миграции
storage/seeds/        — тестовые данные (ручной запуск)
```

## Схема БД

Три таблицы: `manuals`, `tags`, `manual_tags`.

```
manuals (1) ──< manual_tags >── (1) tags
```

| Таблица | Описание |
|---|---|
| `manuals` | Методические материалы (title, author, content, file_path, views_count) |
| `tags` | Справочник тегов (уникальное имя) |
| `manual_tags` | Связь M:N между материалами и тегами |

### Почему три таблицы, а не одна или две

**Плохой вариант — теги строкой в `manuals`:**

```sql
-- антипаттерн
tags = 'go,backend,tutorial'
```

- нарушение 1NF (несколько значений в одном поле);
- сложно искать материалы по тегу;
- дублирование и опечатки (`Go` vs `go`).

**Две таблицы без связующей — тоже не работает:** связь «один материал — много тегов, один тег — у многих материалов» — это **M:N (many-to-many)**. В реляционной модели её нельзя выразить одним внешним ключом между двумя сущностями.

**Три таблицы — нормализованная схема:**

- `manuals` и `tags` хранят сущности независимо (3NF);
- `manual_tags` — junction-таблица для M:N;
- `tags.name UNIQUE` — каждый тег хранится один раз;
- `ON DELETE CASCADE` — при удалении материала или тега связи удаляются автоматически.

### Миграции

Применить по порядку:

```bash
psql -U postgres -d myapp -f storage/migrations/001_create_manuals.sql
psql -U postgres -d myapp -f storage/migrations/002_create_tags.sql
psql -U postgres -d myapp -f storage/migrations/003_create_manual_tags.sql
```

| Файл | Что создаёт |
|---|---|
| `001_create_manuals.sql` | `manuals` — PK `id` (UUID), NOT NULL, DEFAULT для `views_count` и `created_at` |
| `002_create_tags.sql` | `tags` — PK `id` (SERIAL), UNIQUE на `name` |
| `003_create_manual_tags.sql` | `manual_tags` — FK с CASCADE, составной PK `(manual_id, tag_id)` |

### Тестовые данные (вручную)

После миграций:

```bash
psql -U postgres -d myapp -f storage/seeds/001_sample_data.sql
```

Seed добавляет 5 тегов, 4 методички и связи M:N (например, «Введение в Go» → go, tutorial, backend; тег `backend` — у трёх материалов).

Проверка:

```sql
SELECT m.title, array_agg(t.name ORDER BY t.name) AS tags
FROM manuals m
JOIN manual_tags mt ON mt.manual_id = m.id
JOIN tags t ON t.id = mt.tag_id
GROUP BY m.id, m.title;
```

## Запуск

1. Создайте БД PostgreSQL и примените миграции из `storage/migrations/`.
2. Скопируйте `.env` и укажите параметры подключения к БД (`DB_USER`, `DB_PASSWORD`, `DB_NAME` и др.).
3. Запустите сервер:

```bash
go run ./cmd/app
```

## Healthcheck

```bash
curl http://localhost:8080/health
```

Ответ при успешном соединении с БД:

```json
{"status":"ok","database":"ok"}
```

При недоступной БД — HTTP 503:

```json
{"status":"error","database":"unavailable"}
```

## API: Manuals

Единый формат ответов:

```json
{"data": { ... }}
```

Ошибки:

```json
{"error": {"message": "..."}}
```

Ошибка валидации (`400`):

```json
{
  "error": {
    "message": "validation failed",
    "fields": [
      {"field": "title", "message": "is required"},
      {"field": "author", "message": "is required"}
    ]
  }
}
```

### POST /manuals

```bash
curl -X POST http://localhost:8080/manuals \
  -H "Content-Type: application/json" \
  -d '{"title":"Название","author":"Автор","content":"Текст методички"}'
```

Ответ `201`:

```json
{
  "data": {
    "id": "...",
    "title": "Название",
    "author": "Автор",
    "content": "Текст методички",
    "views_count": 0,
    "created_at": "..."
  }
}
```

### GET /manuals/{id}

```bash
curl http://localhost:8080/manuals/{id}
```

Ответ `200` — тот же формат с обёрткой `data`.  
`404` — `{"error":{"message":"manual not found"}}`
