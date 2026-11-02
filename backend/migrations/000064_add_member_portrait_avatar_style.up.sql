ALTER TABLE family_members
  ADD COLUMN portrait_avatar_style VARCHAR(40) NOT NULL DEFAULT 'AUTO' AFTER portrait_glasses_style;
