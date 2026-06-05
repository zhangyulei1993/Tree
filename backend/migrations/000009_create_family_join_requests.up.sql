CREATE TABLE family_join_requests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  applicant_user_id BIGINT NOT NULL,
  applicant_real_name VARCHAR(100) NULL,
  applicant_phone VARCHAR(30) NULL,
  applicant_phone_hash VARCHAR(128) NULL,
  applicant_message VARCHAR(500) NULL,
  request_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  approve_mode VARCHAR(80) NULL,
  bound_member_id BIGINT NULL,
  created_member_id BIGINT NULL,
  handle_result VARCHAR(80) NULL,
  handled_by_user_id BIGINT NULL,
  handled_by_admin_id BIGINT NULL,
  handled_at DATETIME NULL,
  handle_comment VARCHAR(500) NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fjr_family_id (family_id),
  KEY idx_fjr_applicant_user_id (applicant_user_id),
  KEY idx_fjr_request_status (request_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): one user should not create duplicate pending join requests for the same family.
-- TODO(service): approving join requests that create links must run in a transaction.
