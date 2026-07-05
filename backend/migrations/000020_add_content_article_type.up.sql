ALTER TABLE content_articles
    ADD COLUMN content_type VARCHAR(40) NOT NULL DEFAULT 'INTERNAL' AFTER category_key,
    ADD COLUMN external_url VARCHAR(1000) NULL AFTER body,
    ADD KEY idx_content_articles_type_status (content_type, status);
