ALTER TABLE content_articles
    DROP KEY idx_content_articles_type_status,
    DROP COLUMN external_url,
    DROP COLUMN content_type;
