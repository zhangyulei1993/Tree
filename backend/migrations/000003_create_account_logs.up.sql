CREATE TABLE user_phone_history (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  phone VARCHAR(30) NULL,
  phone_hash VARCHAR(128) NULL,
  action_type VARCHAR(80) NOT NULL,
  operator_type VARCHAR(40) NULL,
  operator_user_id BIGINT NULL,
  operator_admin_id BIGINT NULL,
  detail_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_uph_user_id (user_id),
  KEY idx_uph_phone_hash (phone_hash),
  KEY idx_uph_action_type (action_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_account_merge_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  source_user_id BIGINT NOT NULL,
  target_user_id BIGINT NOT NULL,
  merge_reason VARCHAR(500) NULL,
  merge_source VARCHAR(80) NOT NULL,
  member_binding_resolution_json JSON NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  operator_type VARCHAR(40) NULL,
  operator_user_id BIGINT NULL,
  operator_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_uaml_source_user_id (source_user_id),
  KEY idx_uaml_target_user_id (target_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_account_claim_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  claimed_user_id BIGINT NOT NULL,
  temp_user_id BIGINT NULL,
  phone VARCHAR(30) NULL,
  phone_hash VARCHAR(128) NULL,
  claim_via VARCHAR(80) NOT NULL,
  detail_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_uacl_claimed_user_id (claimed_user_id),
  KEY idx_uacl_temp_user_id (temp_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): account merge/claim must run in a database transaction.
-- TODO(service): invalidate source user tokens after account merge.
