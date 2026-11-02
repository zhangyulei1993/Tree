ALTER TABLE families
  ADD COLUMN dissolution_cooldown_until DATETIME NULL AFTER dissolved_at,
  ADD COLUMN dissolution_cooldown_days INT NOT NULL DEFAULT 0 AFTER dissolution_cooldown_until,
  ADD COLUMN dissolution_hidden_at DATETIME NULL AFTER dissolution_cooldown_days,
  ADD COLUMN dissolution_completed_at DATETIME NULL AFTER dissolution_hidden_at;
