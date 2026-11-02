CREATE TABLE user_assistance_codes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  owner_user_id BIGINT NOT NULL,
  code_hash VARCHAR(128) NOT NULL,
  code_status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  expires_at DATETIME NOT NULL,
  replaced_at DATETIME NULL,
  revoked_at DATETIME NULL,
  failed_attempts INT NOT NULL DEFAULT 0,
  locked_at DATETIME NULL,
  session_token_id VARCHAR(128) NULL,
  last_used_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_uac_owner_created (owner_user_id, created_at),
  KEY idx_uac_status_expires (code_status, expires_at),
  UNIQUE KEY uk_uac_code_hash (code_hash),
  CONSTRAINT fk_uac_owner FOREIGN KEY (owner_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
