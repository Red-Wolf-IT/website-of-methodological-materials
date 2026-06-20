CREATE TABLE IF NOT EXISTS manual_tags (
    manual_id UUID NOT NULL REFERENCES manuals (id) ON DELETE CASCADE,
    tag_id    INT NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (manual_id, tag_id)
);
