CREATE TABLE family_public_applications (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  applicant_user_id BIGINT NULL,
  applicant_admin_id BIGINT NULL,
  application_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  application_reason VARCHAR(500) NULL,
  application_snapshot_json JSON NULL,
  review_result VARCHAR(40) NULL,
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fpa_family_id (family_id),
  KEY idx_fpa_status (application_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE visitor_messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  visitor_user_id BIGINT NULL,
  visitor_name VARCHAR(100) NULL,
  visitor_phone VARCHAR(30) NULL,
  visitor_wechat VARCHAR(120) NULL,
  message_content VARCHAR(1000) NOT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  deleted_at DATETIME NULL,
  deleted_by_admin_id BIGINT NULL,
  delete_reason VARCHAR(500) NULL,
  ip VARCHAR(80) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_vm_family_id (family_id),
  KEY idx_vm_status (status),
  KEY idx_vm_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): one family can have only one PENDING public application.
-- TODO(service): PLATFORM_ADMIN cannot review its own public application.
-- TODO(service): public visitor messages must not expose visitor phone or WeChat.
