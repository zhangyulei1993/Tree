ALTER TABLE account_quota_configs
  ADD COLUMN supports_generation_naming TINYINT(1) NOT NULL DEFAULT 0 AFTER max_joined_families;

UPDATE account_quota_configs target
LEFT JOIN account_quota_configs existing_phone_bound
  ON existing_phone_bound.trust_tier = 'PHONE_BOUND'
SET target.trust_tier = 'PHONE_BOUND'
WHERE target.trust_tier = 'PHONE_VERIFIED'
  AND existing_phone_bound.id IS NULL;

UPDATE account_quota_configs
SET supports_generation_naming = 0
WHERE trust_tier = 'WECHAT_ONLY';

UPDATE account_quota_configs
SET supports_generation_naming = 1
WHERE trust_tier = 'PHONE_BOUND';
