CREATE TABLE showcase_display_settings (
  id TINYINT NOT NULL PRIMARY KEY,
  rotation_days SMALLINT NOT NULL DEFAULT 7,
  pinned_showcase_family_id BIGINT NULL,
  pinned_until DATETIME NULL,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT chk_showcase_display_settings_id CHECK (id = 1),
  KEY idx_showcase_display_settings_pinned_until (pinned_until)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO showcase_display_settings (id, rotation_days) VALUES (1, 7);
