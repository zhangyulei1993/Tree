CREATE TABLE launch_splash_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  config_key VARCHAR(64) NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  title VARCHAR(120) NOT NULL,
  tagline VARCHAR(240) NOT NULL,
  duration_ms INT UNSIGNED NOT NULL DEFAULT 3600,
  skip_enabled TINYINT(1) NOT NULL DEFAULT 1,
  updated_by_admin_id BIGINT UNSIGNED NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_launch_splash_configs_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO launch_splash_configs (config_key, enabled, title, tagline, duration_ms, skip_enabled)
VALUES ('HOME', 1, 'Tree', '让每一页家谱，都有来处', 3600, 1);
