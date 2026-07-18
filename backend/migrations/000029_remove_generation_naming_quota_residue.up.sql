DELETE FROM account_feature_overrides
WHERE feature_key = 'GENERATION_NAMING';

ALTER TABLE account_quota_configs
  DROP COLUMN supports_generation_naming;
