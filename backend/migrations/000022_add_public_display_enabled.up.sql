ALTER TABLE families
  ADD COLUMN public_display_enabled TINYINT NOT NULL DEFAULT 0 AFTER public_display_status,
  ADD COLUMN public_enabled_at DATETIME NULL AFTER public_approved_at;

UPDATE families
SET public_display_enabled = 1,
    public_enabled_at = COALESCE(public_approved_at, updated_at)
WHERE public_display_status = 'APPROVED';

CREATE INDEX idx_families_public_display_enabled ON families (public_display_enabled);
