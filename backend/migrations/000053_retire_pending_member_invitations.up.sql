UPDATE family_invitations
SET status = 'CANCELLED',
    cancelled_at = COALESCE(cancelled_at, NOW()),
    cancel_reason = '暂存成员邀请功能已下线，请由家庭管理员从已有成员节点重新邀请',
    updated_at = NOW()
WHERE invite_type = 'JOIN_FAMILY_PENDING_MEMBER'
  AND status = 'PENDING';

DELETE FROM share_configs
WHERE config_key = 'INVITATION_PENDING_MEMBER';
