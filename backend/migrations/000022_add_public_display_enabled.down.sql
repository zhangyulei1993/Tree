DROP INDEX idx_families_public_display_enabled ON families;

ALTER TABLE families
  DROP COLUMN public_enabled_at,
  DROP COLUMN public_display_enabled;
