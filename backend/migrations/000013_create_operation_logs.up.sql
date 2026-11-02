CREATE TABLE operation_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  operator_type VARCHAR(40) NOT NULL,
  operator_admin_id BIGINT NULL,
  operator_user_id BIGINT NULL,
  operator_role VARCHAR(60) NULL,
  module VARCHAR(80) NOT NULL,
  action VARCHAR(100) NOT NULL,
  target_type VARCHAR(80) NULL,
  target_id BIGINT NULL,
  family_id BIGINT NULL,
  member_id BIGINT NULL,
  user_id BIGINT NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  detail_json JSON NULL,
  result VARCHAR(30) NOT NULL DEFAULT 'SUCCESS',
  error_message VARCHAR(500) NULL,
  ip VARCHAR(80) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_operation_logs_operator_admin (operator_admin_id),
  KEY idx_operation_logs_operator_user (operator_user_id),
  KEY idx_operation_logs_module_action (module, action),
  KEY idx_operation_logs_target (target_type, target_id),
  KEY idx_operation_logs_family (family_id),
  KEY idx_operation_logs_member (member_id),
  KEY idx_operation_logs_user (user_id),
  KEY idx_operation_logs_result (result),
  KEY idx_operation_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): all sensitive and important operations must write operation_logs.
-- TODO(service): never log plain password, password hash, verification code, tokens, session_key, AppSecret, full openid, or full unionid.
