ALTER TABLE family_capacity_refunds
  DROP KEY idx_family_capacity_refunds_status,
  DROP COLUMN review_reason,
  DROP COLUMN reviewed_at,
  DROP COLUMN reviewed_by_admin_id;
