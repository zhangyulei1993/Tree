ALTER TABLE family_members
  ADD COLUMN portrait_hair_style VARCHAR(40) NOT NULL DEFAULT 'AUTO' AFTER gender,
  ADD COLUMN portrait_glasses_style VARCHAR(40) NOT NULL DEFAULT 'AUTO' AFTER portrait_hair_style;
