ALTER TABLE family_capacity_refunds
  DROP KEY uk_family_capacity_refunds_order_attempt,
  DROP COLUMN request_reason,
  DROP COLUMN attempt_no,
  ADD UNIQUE KEY uk_family_capacity_refunds_order (order_id);
