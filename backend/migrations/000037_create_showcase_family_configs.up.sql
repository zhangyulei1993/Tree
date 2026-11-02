CREATE TABLE IF NOT EXISTS showcase_family_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  display_name VARCHAR(120) NOT NULL,
  region_text VARCHAR(120) NULL,
  summary VARCHAR(1000) NOT NULL,
  cover_image_url VARCHAR(500) NULL,
  tree_image_url VARCHAR(500) NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
  sort_order INT NOT NULL DEFAULT 0,
  published_at DATETIME NULL,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  KEY idx_showcase_family_configs_status_sort (status, sort_order, id),
  KEY idx_showcase_family_configs_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
