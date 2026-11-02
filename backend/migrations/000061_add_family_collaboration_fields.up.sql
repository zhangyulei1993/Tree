ALTER TABLE families
  ADD COLUMN revision BIGINT NOT NULL DEFAULT 1 AFTER graph_version;

ALTER TABLE family_members
  ADD COLUMN revision BIGINT NOT NULL DEFAULT 1 AFTER status,
  ADD COLUMN profile_source_type VARCHAR(40) NOT NULL DEFAULT 'FAMILY_EDIT' AFTER revision,
  ADD COLUMN profile_verification_status VARCHAR(40) NOT NULL DEFAULT 'PENDING' AFTER profile_source_type,
  ADD COLUMN profile_confirmed_at DATETIME NULL AFTER profile_verification_status,
  ADD COLUMN profile_confirmed_by_user_id BIGINT NULL AFTER profile_confirmed_at,
  ADD KEY idx_family_members_verification (family_id, profile_verification_status, deleted_at);

ALTER TABLE family_relationships
  ADD COLUMN revision BIGINT NOT NULL DEFAULT 1 AFTER status;
