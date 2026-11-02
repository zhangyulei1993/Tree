ALTER TABLE family_invitations
  ADD COLUMN pending_member_label VARCHAR(120) NULL AFTER invite_message;
