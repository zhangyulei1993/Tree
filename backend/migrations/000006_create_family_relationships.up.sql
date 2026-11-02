CREATE TABLE family_relationships (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  from_member_id BIGINT NOT NULL,
  to_member_id BIGINT NOT NULL,
  relationship_type VARCHAR(40) NOT NULL,
  parent_link_type VARCHAR(40) NULL,
  relation_note_type VARCHAR(80) NULL,
  relation_note VARCHAR(500) NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  created_by_user_id BIGINT NULL,
  created_by_admin_id BIGINT NULL,
  deleted_at DATETIME NULL,
  deleted_by_user_id BIGINT NULL,
  deleted_by_admin_id BIGINT NULL,
  delete_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fr_family_id (family_id),
  KEY idx_fr_from_member_id (from_member_id),
  KEY idx_fr_to_member_id (to_member_id),
  KEY idx_fr_relationship_type (relationship_type),
  KEY idx_fr_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TODO(service): relationship_type only allows PARENT_CHILD and SPOUSE; never persist SIBLING.
-- TODO(service): ADD_SIBLING must be implemented through shared parent-child relationships.
-- TODO(service): enforce one PRIMARY father and one PRIMARY mother per child.
-- TODO(service): relationship changes must increment families.graph_version.
