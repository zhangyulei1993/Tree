CREATE TABLE IF NOT EXISTS showcase_family_members (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  showcase_family_id BIGINT NOT NULL,
  display_name VARCHAR(80) NOT NULL,
  gender VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN',
  birth_year SMALLINT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  KEY idx_showcase_family_members_family (showcase_family_id, sort_order, id),
  KEY idx_showcase_family_members_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS showcase_family_relationships (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  showcase_family_id BIGINT NOT NULL,
  from_member_id BIGINT NOT NULL,
  to_member_id BIGINT NOT NULL,
  relationship_type VARCHAR(20) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  UNIQUE KEY uk_showcase_family_relationships_active (showcase_family_id, from_member_id, to_member_id, relationship_type, deleted_at),
  KEY idx_showcase_family_relationships_family (showcase_family_id, relationship_type, id),
  KEY idx_showcase_family_relationships_from (from_member_id),
  KEY idx_showcase_family_relationships_to (to_member_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
