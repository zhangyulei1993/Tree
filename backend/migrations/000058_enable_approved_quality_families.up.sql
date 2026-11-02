UPDATE families
SET public_display_enabled = 1,
    searchable = 1,
    public_enabled_at = COALESCE(public_enabled_at, NOW())
WHERE deleted_at IS NULL
  AND status = 'NORMAL'
  AND public_display_status = 'APPROVED'
  AND public_display_enabled = 0;
