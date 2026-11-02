ALTER TABLE family_invitation_reward_referrals
  DROP KEY idx_family_invitation_reward_referrals_inviter_active,
  DROP COLUMN expires_at;
