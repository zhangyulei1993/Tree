CREATE TABLE backup_operation_configs (
  id TINYINT UNSIGNED NOT NULL,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  retention_days INT NOT NULL DEFAULT 14,
  max_storage_mb INT NOT NULL DEFAULT 512,
  max_disk_usage_percent INT NOT NULL DEFAULT 85,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  CONSTRAINT fk_backup_operation_configs_admin FOREIGN KEY (updated_by_admin_id) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO backup_operation_configs (id, enabled, retention_days, max_storage_mb, max_disk_usage_percent)
VALUES (1, 1, 14, 512, 85);
