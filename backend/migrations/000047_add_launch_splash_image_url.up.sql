ALTER TABLE launch_splash_configs
  ADD COLUMN image_url VARCHAR(500) NOT NULL DEFAULT '' AFTER tagline;
