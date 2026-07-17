DELETE target
FROM account_feature_overrides target
JOIN account_feature_overrides source
  ON source.feature_key = 'GENERATION_NAMING'
 AND source.phone_hash = target.phone_hash
WHERE target.feature_key = 'FEATURE_PREVIEW';

ALTER TABLE account_quota_configs
  DROP COLUMN supports_feature_preview;
