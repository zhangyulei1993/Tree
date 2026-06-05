CREATE TABLE family_member_user_links (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  member_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  link_status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  link_source VARCHAR(80) NOT NULL,
  family_role VARCHAR(40) NOT NULL DEFAULT 'MEMBER',
  invitation_id BIGINT NULL,
  join_request_id BIGINT NULL,
  role_granted_at DATETIME NULL,
  role_granted_by_user_id BIGINT NULL,
  role_granted_by_admin_id BIGINT NULL,
  unlinked_at DATETIME NULL,
  unlinked_reason VARCHAR(500) NULL,
  unlinked_by_user_id BIGINT NULL,
  unlinked_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fmul_family_id (family_id),
  KEY idx_fmul_member_id (member_id),
  KEY idx_fmul_user_id (user_id),
  KEY idx_fmul_link_status (link_status),
  KEY idx_fmul_family_role (family_role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): one user can have only one ACTIVE member link in the same family.
-- TODO(service): one member can have only one ACTIVE user link.
-- TODO(service): one family can have only one FOUNDER.
