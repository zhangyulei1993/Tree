ALTER TABLE account_quota_configs
  ADD COLUMN supports_feature_preview TINYINT(1) NOT NULL DEFAULT 0 AFTER supports_generation_naming;

UPDATE account_quota_configs
SET supports_feature_preview = supports_generation_naming;

INSERT INTO account_feature_overrides (feature_key, phone_hash, phone_mask, updated_by_admin_id, created_at, updated_at)
SELECT 'FEATURE_PREVIEW', source.phone_hash, source.phone_mask, source.updated_by_admin_id, source.created_at, source.updated_at
FROM account_feature_overrides source
LEFT JOIN account_feature_overrides existing
  ON existing.feature_key = 'FEATURE_PREVIEW'
 AND existing.phone_hash = source.phone_hash
WHERE source.feature_key = 'GENERATION_NAMING'
  AND existing.id IS NULL;
