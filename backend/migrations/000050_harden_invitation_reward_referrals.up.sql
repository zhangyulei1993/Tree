ALTER TABLE family_invitation_reward_referrals
  ADD COLUMN expires_at DATETIME NULL AFTER qualified_at,
  ADD KEY idx_family_invitation_reward_referrals_inviter_active (family_id, inviter_user_id, status, expires_at);

UPDATE family_invitation_reward_referrals
SET expires_at = DATE_ADD(created_at, INTERVAL 7 DAY)
WHERE status = 'PENDING' AND expires_at IS NULL;
