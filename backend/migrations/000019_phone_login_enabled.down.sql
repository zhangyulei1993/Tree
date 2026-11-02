DROP INDEX uk_users_phone_hash ON users;

UPDATE account_quota_configs
SET trust_tier = 'PHONE_VERIFIED'
WHERE trust_tier = 'PHONE_BOUND';

ALTER TABLE users
  DROP COLUMN phone_login_enabled;
