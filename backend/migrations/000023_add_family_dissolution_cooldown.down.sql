ALTER TABLE families
  DROP COLUMN dissolution_completed_at,
  DROP COLUMN dissolution_hidden_at,
  DROP COLUMN dissolution_cooldown_days,
  DROP COLUMN dissolution_cooldown_until;
