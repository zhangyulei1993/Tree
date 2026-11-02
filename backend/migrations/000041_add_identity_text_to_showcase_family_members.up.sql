ALTER TABLE showcase_family_members
  ADD COLUMN identity_text VARCHAR(120) NOT NULL DEFAULT '' AFTER display_name;
