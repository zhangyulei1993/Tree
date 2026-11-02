ALTER TABLE family_members
  ADD COLUMN recovery_discarded_at DATETIME NULL AFTER delete_reason,
  ADD COLUMN recovery_discarded_by_user_id BIGINT NULL AFTER recovery_discarded_at,
  ADD KEY idx_family_members_recovery (family_id, status, deleted_at, recovery_discarded_at);
