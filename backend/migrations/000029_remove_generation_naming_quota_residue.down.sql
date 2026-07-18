ALTER TABLE account_quota_configs
  ADD COLUMN supports_generation_naming TINYINT(1) NOT NULL DEFAULT 0 AFTER max_joined_families;

UPDATE account_quota_configs
SET supports_generation_naming = CASE
  WHEN trust_tier = 'PHONE_BOUND' THEN 1
  ELSE 0
END;
