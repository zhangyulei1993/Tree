CREATE TABLE family_founder_transfer_requests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  from_member_id BIGINT NOT NULL,
  from_user_id BIGINT NOT NULL,
  to_member_id BIGINT NOT NULL,
  to_user_id BIGINT NOT NULL,
  request_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  request_reason VARCHAR(500) NULL,
  review_result VARCHAR(40) NULL,
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  completed_at DATETIME NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fftr_family_id (family_id),
  KEY idx_fftr_status (request_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_dissolution_requests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  requester_member_id BIGINT NOT NULL,
  requester_user_id BIGINT NOT NULL,
  request_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  request_reason VARCHAR(500) NULL,
  request_snapshot_json JSON NULL,
  review_result VARCHAR(40) NULL,
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  completed_at DATETIME NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fdr_family_id (family_id),
  KEY idx_fdr_status (request_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): one family can have only one PENDING founder transfer request.
-- TODO(service): one family can have only one PENDING dissolution request.
-- TODO(service): PLATFORM_ADMIN cannot approve founder transfer or family dissolution.
