-- Тестовые данные для ручного наполнения БД.
-- Запускать ПОСЛЕ миграций из storage/migrations/.
--
-- psql -U postgres -d myapp -f storage/seeds/001_sample_data.sql

INSERT INTO tags (name) VALUES
    ('go'),
    ('backend'),
    ('tutorial'),
    ('database'),
    ('devops')
ON CONFLICT (name) DO NOTHING;

INSERT INTO manuals (id, title, author, content, file_path, views_count) VALUES
    (
        'a1000000-0000-4000-8000-000000000001',
        'Введение в Go',
        'Иванов И.И.',
        'Краткий обзор синтаксиса Go: переменные, функции, структуры и пакеты.',
        NULL,
        42
    ),
    (
        'a1000000-0000-4000-8000-000000000002',
        'REST API на chi',
        'Петров П.П.',
        'Пошаговое руководство по созданию HTTP-сервиса с роутером chi.',
        '/uploads/rest-api-chi.pdf',
        18
    ),
    (
        'a1000000-0000-4000-8000-000000000003',
        'PostgreSQL для начинающих',
        'Сидорова А.А.',
        'Основы SQL, индексы, внешние ключи и транзакции в PostgreSQL.',
        NULL,
        7
    ),
    (
        'a1000000-0000-4000-8000-000000000004',
        'Docker в разработке',
        'Козлов Д.Д.',
        'Контейнеризация приложений: Dockerfile, docker-compose, best practices.',
        '/uploads/docker-guide.pdf',
        25
    )
ON CONFLICT (id) DO NOTHING;

-- M:N: один материал — несколько тегов, один тег — у нескольких материалов
INSERT INTO manual_tags (manual_id, tag_id)
SELECT m.id, t.id
FROM manuals m
CROSS JOIN tags t
WHERE (m.title, t.name) IN (
    ('Введение в Go',           'go'),
    ('Введение в Go',           'tutorial'),
    ('Введение в Go',           'backend'),
    ('REST API на chi',         'go'),
    ('REST API на chi',         'backend'),
    ('REST API на chi',         'tutorial'),
    ('PostgreSQL для начинающих', 'database'),
    ('PostgreSQL для начинающих', 'tutorial'),
    ('Docker в разработке',     'devops'),
    ('Docker в разработке',     'backend')
)
ON CONFLICT (manual_id, tag_id) DO NOTHING;
