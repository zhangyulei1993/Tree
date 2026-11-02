ALTER TABLE launch_splash_configs
  ADD COLUMN day_image_url VARCHAR(500) NOT NULL DEFAULT '' AFTER image_url,
  ADD COLUMN night_image_url VARCHAR(500) NOT NULL DEFAULT '' AFTER day_image_url;
