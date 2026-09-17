ALTER TABLE share_configs
  ADD COLUMN image_urls TEXT NULL AFTER image_url;

UPDATE share_configs
SET image_urls = JSON_ARRAY(image_url)
WHERE image_urls IS NULL OR image_urls = '';
