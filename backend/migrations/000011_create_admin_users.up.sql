CREATE TABLE admin_users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(100) NULL,
  phone VARCHAR(30) NULL,
  email VARCHAR(150) NULL,
  role VARCHAR(40) NOT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  failed_login_count INT NOT NULL DEFAULT 0,
  locked_until DATETIME NULL,
  last_login_at DATETIME NULL,
  last_login_ip VARCHAR(80) NULL,
  password_changed_at DATETIME NULL,
  disabled_at DATETIME NULL,
  disabled_by_admin_id BIGINT NULL,
  disabled_reason VARCHAR(500) NULL,
  deleted_at DATETIME NULL,
  deleted_by_admin_id BIGINT NULL,
  unlocked_at DATETIME NULL,
  unlocked_by_admin_id BIGINT NULL,
  remark VARCHAR(500) NULL,
  created_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_admin_users_username (username),
  KEY idx_admin_users_role (role),
  KEY idx_admin_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): ROOT_ADMIN must be unique and cannot be managed by lower roles.
-- TODO(service): admin passwords must always be bcrypt hashes, never plain text.
