ALTER TABLE family_invitation_reward_tiers
  DROP KEY idx_family_invitation_reward_tiers_active,
  DROP COLUMN retired;
