ALTER TABLE family_invitation_reward_tiers
  ADD COLUMN retired TINYINT(1) NOT NULL DEFAULT 0 AFTER sort_order,
  ADD KEY idx_family_invitation_reward_tiers_active (config_id, retired, sort_order);
