UPDATE content_articles
SET
    body = REPLACE(body, '主入口是微信快捷登录，页面底部的小字是手机号备用登录。第一次使用一般选择微信登录即可。', '请使用微信快捷登录。需要发起邀请时，系统会引导你授权绑定微信手机号。'),
    updated_at = UTC_TIMESTAMP()
WHERE slug = 'tree-new-user-picture-guide';
