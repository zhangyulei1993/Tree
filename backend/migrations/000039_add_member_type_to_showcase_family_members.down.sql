DROP INDEX idx_showcase_family_members_type ON showcase_family_members;

ALTER TABLE showcase_family_members
  DROP COLUMN member_type;
