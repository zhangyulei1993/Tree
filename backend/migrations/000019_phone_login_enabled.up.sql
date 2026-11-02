-- Pre-check (run manually on version 18 before migration):
-- 1) Duplicate phone_hash (must match uk_users_phone_hash unique index scope):
-- SELECT phone_hash, COUNT(*) AS c
-- FROM users
-- WHERE phone_hash IS NOT NULL
-- GROUP BY phone_hash
-- HAVING c > 1;
-- 2) Active users with both phone_hash and password_hash (will be auto-enabled; no phone_login_enabled column required):
-- SELECT id
-- FROM users
-- WHERE deleted_at IS NULL
--   AND status = 'ACTIVE'
--   AND phone_hash IS NOT NULL
--   AND phone_hash <> ''
--   AND password_hash IS NOT NULL
--   AND password_hash <> '';

ALTER TABLE users
  ADD COLUMN phone_login_enabled TINYINT NOT NULL DEFAULT 0 AFTER phone_verified;

UPDATE users
SET phone_login_enabled = 1
WHERE deleted_at IS NULL
  AND status = 'ACTIVE'
  AND phone_hash IS NOT NULL
  AND phone_hash <> ''
  AND password_hash IS NOT NULL
  AND password_hash <> '';

UPDATE account_quota_configs
SET trust_tier = 'PHONE_BOUND'
WHERE trust_tier = 'PHONE_VERIFIED';

CREATE UNIQUE INDEX uk_users_phone_hash ON users (phone_hash);

-- Post-check (run after migration):
-- SELECT id
-- FROM users
-- WHERE phone_login_enabled = 1
--   AND (
--     phone_hash IS NULL OR phone_hash = ''
--     OR password_hash IS NULL OR password_hash = ''
--   );
