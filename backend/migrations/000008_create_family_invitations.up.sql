CREATE TABLE family_invitations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  target_member_id BIGINT NOT NULL,
  inviter_user_id BIGINT NULL,
  inviter_admin_id BIGINT NULL,
  target_user_id BIGINT NULL,
  target_phone VARCHAR(30) NULL,
  target_phone_hash VARCHAR(128) NULL,
  invite_type VARCHAR(80) NOT NULL,
  invite_channel VARCHAR(40) NOT NULL,
  invite_actor_type VARCHAR(80) NOT NULL,
  family_role_after_accept VARCHAR(40) NOT NULL DEFAULT 'MEMBER',
  invite_token VARCHAR(180) NULL,
  invite_message VARCHAR(500) NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  accepted_by_user_id BIGINT NULL,
  accepted_at DATETIME NULL,
  rejected_by_user_id BIGINT NULL,
  rejected_at DATETIME NULL,
  reject_reason VARCHAR(500) NULL,
  cancelled_by_user_id BIGINT NULL,
  cancelled_by_admin_id BIGINT NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  expired_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fi_family_id (family_id),
  KEY idx_fi_target_member_id (target_member_id),
  KEY idx_fi_invite_token (invite_token),
  KEY idx_fi_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): one member can have only one PENDING invitation.
-- TODO(service): PLATFORM_ADMIN cannot generate share invitation links.
-- TODO(service): accepting invitation must not increment graph_version.
