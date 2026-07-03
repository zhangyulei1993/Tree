-- Enforce one ACTIVE WeChat identity per provider/provider_app_id/openid_hash.
-- CANCELLED/MERGED historical rows may repeat the same openid_hash.
--
-- Pre-check (must match active_identity_key generated column rule):
-- SELECT active_identity_key, COUNT(*) AS cnt
-- FROM (
--   SELECT
--     CASE
--       WHEN identity_status = 'ACTIVE' AND deleted_at IS NULL AND openid_hash IS NOT NULL THEN
--         SHA2(CONCAT(COALESCE(provider, ''), CHAR(0), COALESCE(provider_app_id, ''), CHAR(0), openid_hash), 256)
--       ELSE NULL
--     END AS active_identity_key
--   FROM user_auth_identities
-- ) AS duplicates
-- WHERE active_identity_key IS NOT NULL
-- GROUP BY active_identity_key
-- HAVING cnt > 1;

ALTER TABLE user_auth_identities
  ADD COLUMN active_identity_key CHAR(64) AS (
    CASE
      WHEN identity_status = 'ACTIVE' AND deleted_at IS NULL AND openid_hash IS NOT NULL THEN
        SHA2(CONCAT(COALESCE(provider, ''), CHAR(0), COALESCE(provider_app_id, ''), CHAR(0), openid_hash), 256)
      ELSE NULL
    END
  ) STORED NULL COMMENT 'ACTIVE identity dedupe key; NULL for non-ACTIVE rows';

CREATE UNIQUE INDEX uk_uai_active_identity_key
  ON user_auth_identities (active_identity_key);
