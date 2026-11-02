CREATE TABLE IF NOT EXISTS share_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  config_key VARCHAR(64) NOT NULL,
  title_template VARCHAR(300) NOT NULL,
  image_url VARCHAR(500) NOT NULL,
  description VARCHAR(500) NULL,
  updated_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_share_configs_key (config_key),
  KEY idx_share_configs_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO share_configs (config_key, title_template, image_url, description)
VALUES
  ('HOME', '家脉｜记录家族，连接亲人', 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', '首页分享'),
  ('PUBLIC_FAMILY', '{{familyName}}｜公开家庭主页', 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', '公开家庭分享，支持变量：{{familyName}}'),
  ('INVITATION_NODE', '{{familyName}}邀请你确认「{{targetMemberName}}」身份并加入家庭树', 'https://tapi.bigbigboy.cn/api/static/content/share/family-invitation.jpg', '节点绑定邀请，支持变量：{{familyName}}、{{targetMemberName}}'),
  ('INVITATION_PENDING_MEMBER', '{{familyName}}邀请你加入家庭（{{targetMemberName}}）', 'https://tapi.bigbigboy.cn/api/static/content/share/family-invitation.jpg', '暂存成员邀请，支持变量：{{familyName}}、{{targetMemberName}}')
ON DUPLICATE KEY UPDATE config_key = VALUES(config_key);
