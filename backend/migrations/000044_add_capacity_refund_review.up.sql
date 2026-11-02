ALTER TABLE family_capacity_refunds
  ADD COLUMN reviewed_by_admin_id BIGINT NULL AFTER operator_user_id,
  ADD COLUMN reviewed_at DATETIME NULL AFTER requested_at,
  ADD COLUMN review_reason VARCHAR(500) NULL AFTER failed_at,
  ADD KEY idx_family_capacity_refunds_status (status, requested_at);
