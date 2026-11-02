CREATE TABLE family_invitation_reward_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  activation_remaining_slots INT NOT NULL DEFAULT 1,
  max_reward_members INT NOT NULL DEFAULT 3,
  require_phone_bound TINYINT(1) NOT NULL DEFAULT 1,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_invitation_reward_tiers (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  config_id BIGINT UNSIGNED NOT NULL,
  required_invites INT NOT NULL,
  cumulative_reward_members INT NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_family_invitation_reward_tiers_config_required (config_id, required_invites),
  CONSTRAINT fk_family_invitation_reward_tiers_config FOREIGN KEY (config_id) REFERENCES family_invitation_reward_configs(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_invitation_reward_referrals (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  inviter_user_id BIGINT NOT NULL,
  referral_code VARCHAR(64) NOT NULL,
  invitee_user_id BIGINT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'PENDING',
  qualified_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_family_invitation_reward_referrals_code (referral_code),
  UNIQUE KEY uk_family_invitation_reward_referrals_invitee (invitee_user_id),
  KEY idx_family_invitation_reward_referrals_family_status (family_id, status),
  CONSTRAINT fk_family_invitation_reward_referrals_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_invitation_reward_referrals_inviter FOREIGN KEY (inviter_user_id) REFERENCES users(id),
  CONSTRAINT fk_family_invitation_reward_referrals_invitee FOREIGN KEY (invitee_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_invitation_reward_grants (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  tier_id BIGINT UNSIGNED NOT NULL,
  added_members INT NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE',
  granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_family_invitation_reward_grants_family_tier (family_id, tier_id),
  KEY idx_family_invitation_reward_grants_family_status (family_id, status),
  CONSTRAINT fk_family_invitation_reward_grants_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_invitation_reward_grants_tier FOREIGN KEY (tier_id) REFERENCES family_invitation_reward_tiers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO family_invitation_reward_configs (id, enabled, activation_remaining_slots, max_reward_members, require_phone_bound)
VALUES (1, 1, 1, 3, 1);
INSERT INTO family_invitation_reward_tiers (config_id, required_invites, cumulative_reward_members, sort_order)
VALUES (1, 6, 1, 10), (1, 10, 2, 20), (1, 18, 3, 30);
