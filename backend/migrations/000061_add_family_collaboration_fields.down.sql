ALTER TABLE family_relationships DROP COLUMN revision;

ALTER TABLE family_members
  DROP INDEX idx_family_members_verification,
  DROP COLUMN profile_confirmed_by_user_id,
  DROP COLUMN profile_confirmed_at,
  DROP COLUMN profile_verification_status,
  DROP COLUMN profile_source_type,
  DROP COLUMN revision;

ALTER TABLE families DROP COLUMN revision;
