INSERT INTO admin_users (
  username,
  password_hash,
  display_name,
  role,
  status,
  remark,
  created_at,
  updated_at
) VALUES (
  'admin',
  '$2a$10$7EqJtq98hPqEX7fNZaFWoOhiMEqvN24I37n44hIWqL0/4Q2QKlQpi',
  'ROOT_ADMIN',
  'ROOT_ADMIN',
  'ACTIVE',
  'Seeded ROOT_ADMIN with bcrypt hash placeholder; replace before production use.',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
);

-- TODO(service): enforce ROOT_ADMIN uniqueness before creating any additional admin users.
-- TODO(ops): replace this bcrypt hash during deployment; do not use or document a plain default password.
