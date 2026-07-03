CREATE TABLE account_quota_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  trust_tier VARCHAR(40) NOT NULL,
  max_owned_families INT NOT NULL,
  max_members_per_owned_family INT NOT NULL,
  max_joined_families INT NOT NULL,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_account_quota_configs_trust_tier (trust_tier)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO account_quota_configs (
  trust_tier,
  max_owned_families,
  max_members_per_owned_family,
  max_joined_families
) VALUES
  ('WECHAT_ONLY', 1, 10, 1),
  ('PHONE_VERIFIED', 1, 20, 5);
