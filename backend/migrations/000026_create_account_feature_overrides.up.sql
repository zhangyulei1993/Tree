CREATE TABLE account_feature_overrides (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  feature_key VARCHAR(80) NOT NULL,
  phone_hash VARCHAR(128) NOT NULL,
  phone_mask VARCHAR(30) NOT NULL,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_account_feature_overrides_feature_phone (feature_key, phone_hash),
  KEY idx_account_feature_overrides_feature_key (feature_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
