INSERT INTO share_configs (config_key, title_template, image_url, description)
VALUES
  ('PUBLIC_FAMILY_TIMELINE', '{{familyName}}｜优质家庭案例', 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', '优质家庭朋友圈分享，支持变量：{{familyName}}'),
  ('INVITATION_REWARD', '邀请你帮{{familyName}}助力解锁名额', 'https://tapi.bigbigboy.cn/api/static/content/share/family-invitation-v2.jpg', '家庭邀请助力，支持变量：{{familyName}}'),
  ('INVITATION_DETAIL', '{{familyName}}邀请你确认家庭成员身份', 'https://tapi.bigbigboy.cn/api/static/content/share/family-invitation-v2.jpg', '邀请确认页转发，支持变量：{{familyName}}'),
  ('CONTENT_ARTICLE', '{{articleTitle}}｜家脉阅读', 'https://tapi.bigbigboy.cn/api/static/content/share/mini-program-default.jpg', '文章内容分享，支持变量：{{articleTitle}}')
ON DUPLICATE KEY UPDATE
  title_template = VALUES(title_template),
  description = VALUES(description);
