DROP INDEX uk_uai_active_identity_key ON user_auth_identities;

ALTER TABLE user_auth_identities
  DROP COLUMN active_identity_key;
