ALTER TABLE family_members
  DROP INDEX idx_family_members_recovery,
  DROP COLUMN recovery_discarded_by_user_id,
  DROP COLUMN recovery_discarded_at;
