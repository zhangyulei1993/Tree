ALTER TABLE family_capacity_refunds
  DROP KEY uk_family_capacity_refunds_order,
  ADD COLUMN attempt_no INT NOT NULL DEFAULT 1 AFTER order_id,
  ADD COLUMN request_reason VARCHAR(500) NULL AFTER requested_at,
  ADD UNIQUE KEY uk_family_capacity_refunds_order_attempt (order_id, attempt_no);
