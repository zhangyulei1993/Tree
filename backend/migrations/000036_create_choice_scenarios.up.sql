CREATE TABLE choice_scenarios (
  id BIGINT NOT NULL AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  title VARCHAR(60) NOT NULL,
  options_json JSON NOT NULL,
  status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
  use_count INT NOT NULL DEFAULT 0,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_choice_scenarios_feed (status, expires_at, created_at),
  KEY idx_choice_scenarios_user_created (user_id, created_at),
  CONSTRAINT fk_choice_scenarios_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
