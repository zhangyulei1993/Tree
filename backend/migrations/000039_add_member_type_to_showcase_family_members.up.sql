ALTER TABLE showcase_family_members
  ADD COLUMN member_type VARCHAR(40) NOT NULL DEFAULT 'LINEAGE_MEMBER' AFTER showcase_family_id;

CREATE INDEX idx_showcase_family_members_type
  ON showcase_family_members (showcase_family_id, member_type, deleted_at);
