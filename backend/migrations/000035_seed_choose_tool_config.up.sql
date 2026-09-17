INSERT INTO tool_configs (tool_key, display_name, description, visible, enabled, pinned, highlighted, sort_order)
VALUES ('CHOOSE', '该选什么', '把选项交给随机选择，帮你快速做决定', 1, 0, 0, 0, 0)
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  description = VALUES(description);
