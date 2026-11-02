CREATE TABLE user_auth_identities (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  provider VARCHAR(80) NOT NULL,
  provider_app_id VARCHAR(120) NULL,
  openid VARCHAR(180) NULL,
  openid_hash VARCHAR(128) NULL,
  unionid VARCHAR(180) NULL,
  unionid_hash VARCHAR(128) NULL,
  identity_status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  bound_at DATETIME NULL,
  unbound_at DATETIME NULL,
  last_login_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  KEY idx_uai_user_id (user_id),
  KEY idx_uai_provider_openid (provider, provider_app_id, openid_hash),
  KEY idx_uai_unionid (unionid_hash),
  KEY idx_uai_status (identity_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE verification_codes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  phone VARCHAR(30) NOT NULL,
  phone_hash VARCHAR(128) NOT NULL,
  code_hash VARCHAR(255) NOT NULL,
  scene VARCHAR(80) NOT NULL,
  client_type VARCHAR(80) NOT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  expired_at DATETIME NOT NULL,
  used_at DATETIME NULL,
  ip VARCHAR(80) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_vc_phone_scene (phone_hash, scene),
  KEY idx_vc_status (status),
  KEY idx_vc_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): enforce one ACTIVE identity per provider/provider_app_id/openid_hash.
-- TODO(service): never store or log plain verification codes.
