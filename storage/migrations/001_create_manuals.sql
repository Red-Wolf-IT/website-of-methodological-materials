CREATE TABLE IF NOT EXISTS manuals (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    author      VARCHAR(255) NOT NULL,
    content     TEXT NOT NULL,
    file_path   VARCHAR(512),
    views_count INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP
);
