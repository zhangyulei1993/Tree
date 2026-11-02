CREATE TABLE quality_family_eligibility_configs (
  id BIGINT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  min_member_count INT NOT NULL DEFAULT 3,
  min_relationship_count INT NOT NULL DEFAULT 2,
  min_bound_member_count INT NOT NULL DEFAULT 1,
  min_profile_complete_rate INT NOT NULL DEFAULT 0,
  min_bound_member_rate INT NOT NULL DEFAULT 0,
  require_description BOOLEAN NOT NULL DEFAULT FALSE,
  require_location BOOLEAN NOT NULL DEFAULT FALSE,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  CONSTRAINT fk_quality_family_config_admin FOREIGN KEY (updated_by_admin_id) REFERENCES admin_users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO quality_family_eligibility_configs (id, enabled, min_member_count, min_relationship_count, min_bound_member_count, min_profile_complete_rate, min_bound_member_rate, require_description, require_location)
VALUES (1, TRUE, 3, 2, 1, 0, 0, FALSE, FALSE);
