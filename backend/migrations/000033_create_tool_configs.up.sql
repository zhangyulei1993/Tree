CREATE TABLE IF NOT EXISTS tool_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  tool_key VARCHAR(80) NOT NULL,
  display_name VARCHAR(120) NOT NULL,
  description VARCHAR(300) NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_tool_configs_key (tool_key),
  KEY idx_tool_configs_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO tool_configs (tool_key, display_name, description, enabled)
VALUES
  ('KINSHIP_QUERY', '称谓查询', '按关系、性别和长幼查询称谓', 1),
  ('TRADITIONAL_FESTIVALS', '传统节日', '查看传统节日、法定假期与节日文案', 1),
  ('SOLAR_TERMS', '农历节气', '查看二十四节气与时令变化', 1),
  ('FAMILY_STORIES', '家庭事迹', '查看家庭成员与关系变动记录', 1)
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  description = VALUES(description);
