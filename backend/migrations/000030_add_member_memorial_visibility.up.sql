ALTER TABLE family_members
  ADD COLUMN memorial_visible TINYINT(1) NOT NULL DEFAULT 1 AFTER lineage_note;
