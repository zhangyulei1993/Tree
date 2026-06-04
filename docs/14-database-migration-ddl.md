# 14. 数据库 Migration / DDL 设计 v1

> 数据库：MySQL 8。字符集：utf8mb4。状态字段使用 VARCHAR，不使用 MySQL ENUM。Codex 应将本文拆分为 `backend/migrations/*.sql`。

## 1. 通用字段

核心表默认包含：

```sql
created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
```

软删除表包含：

```sql
deleted_at DATETIME NULL
```

## 2. users

```sql
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  phone VARCHAR(30) NULL,
  phone_hash VARCHAR(128) NULL,
  phone_verified TINYINT NOT NULL DEFAULT 0,
  password_hash VARCHAR(255) NULL,
  nickname VARCHAR(100) NULL,
  real_name VARCHAR(100) NULL,
  avatar_url VARCHAR(500) NULL,
  account_origin VARCHAR(80) NOT NULL,
  register_client VARCHAR(80) NOT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  merged_to_user_id BIGINT NULL,
  merged_at DATETIME NULL,
  claimed_at DATETIME NULL,
  claimed_via VARCHAR(80) NULL,
  remark VARCHAR(500) NULL,
  last_login_at DATETIME NULL,
  last_login_ip VARCHAR(80) NULL,
  last_login_client VARCHAR(80) NULL,
  disabled_at DATETIME NULL,
  disabled_by_admin_id BIGINT NULL,
  disabled_reason VARCHAR(500) NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  KEY idx_users_phone (phone),
  KEY idx_users_phone_hash (phone_hash),
  KEY idx_users_status (status),
  KEY idx_users_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

业务约束：同一时间 phone 不能被多个 ACTIVE / PENDING_CLAIM user 占用；CANCELLED 用户 phone = NULL。

## 3. user_auth_identities

```sql
CREATE TABLE user_auth_identities (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  provider VARCHAR(80) NOT NULL,
  provider_app_id VARCHAR(120) NULL,
  openid VARCHAR(180) NULL,
  openid_hash VARCHAR(128) NULL,
  unionid VARCHAR(180) NULL,
  unionid_hash VARCHAR(128) NULL,
  identity_status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  bound_at DATETIME NULL,
  unbound_at DATETIME NULL,
  last_login_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  KEY idx_uai_user_id (user_id),
  KEY idx_uai_provider_openid (provider, provider_app_id, openid_hash),
  KEY idx_uai_unionid (unionid_hash),
  KEY idx_uai_status (identity_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 4. verification_codes

```sql
CREATE TABLE verification_codes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  phone VARCHAR(30) NOT NULL,
  phone_hash VARCHAR(128) NOT NULL,
  code_hash VARCHAR(255) NOT NULL,
  scene VARCHAR(80) NOT NULL,
  client_type VARCHAR(80) NOT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  expired_at DATETIME NOT NULL,
  used_at DATETIME NULL,
  ip VARCHAR(80) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_vc_phone_scene (phone_hash, scene),
  KEY idx_vc_status (status),
  KEY idx_vc_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 5. account logs

```sql
CREATE TABLE user_phone_history (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  phone VARCHAR(30) NULL,
  phone_hash VARCHAR(128) NULL,
  action_type VARCHAR(80) NOT NULL,
  operator_type VARCHAR(40) NULL,
  operator_user_id BIGINT NULL,
  operator_admin_id BIGINT NULL,
  detail_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_uph_user_id (user_id),
  KEY idx_uph_phone_hash (phone_hash),
  KEY idx_uph_action_type (action_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_account_merge_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  source_user_id BIGINT NOT NULL,
  target_user_id BIGINT NOT NULL,
  merge_reason VARCHAR(500) NULL,
  merge_source VARCHAR(80) NOT NULL,
  member_binding_resolution_json JSON NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  operator_type VARCHAR(40) NULL,
  operator_user_id BIGINT NULL,
  operator_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_uaml_source_user_id (source_user_id),
  KEY idx_uaml_target_user_id (target_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_account_claim_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  claimed_user_id BIGINT NOT NULL,
  temp_user_id BIGINT NULL,
  phone VARCHAR(30) NULL,
  phone_hash VARCHAR(128) NULL,
  claim_via VARCHAR(80) NOT NULL,
  detail_json JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_uacl_claimed_user_id (claimed_user_id),
  KEY idx_uacl_temp_user_id (temp_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 6. families

```sql
CREATE TABLE families (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_name VARCHAR(150) NOT NULL,
  family_surname VARCHAR(40) NOT NULL,
  native_place VARCHAR(200) NULL,
  region_code VARCHAR(50) NULL,
  region_text VARCHAR(200) NULL,
  description TEXT NULL,
  avatar_url VARCHAR(500) NULL,
  creator_user_id BIGINT NULL,
  current_founder_member_id BIGINT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'NORMAL',
  searchable TINYINT NOT NULL DEFAULT 1,
  public_display_status VARCHAR(40) NOT NULL DEFAULT 'PRIVATE',
  public_contact_name VARCHAR(100) NULL,
  public_contact_phone VARCHAR(30) NULL,
  public_contact_wechat VARCHAR(120) NULL,
  public_contact_note VARCHAR(500) NULL,
  public_contact_visible TINYINT NOT NULL DEFAULT 1,
  tree_mode VARCHAR(40) NOT NULL DEFAULT 'LIST_TREE',
  graph_version BIGINT NOT NULL DEFAULT 1,
  public_applied_at DATETIME NULL,
  public_approved_at DATETIME NULL,
  public_taken_down_at DATETIME NULL,
  disabled_at DATETIME NULL,
  disabled_by_admin_id BIGINT NULL,
  disabled_reason VARCHAR(500) NULL,
  dissolved_at DATETIME NULL,
  restored_at DATETIME NULL,
  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_families_name (family_name),
  KEY idx_families_surname (family_surname),
  KEY idx_families_status (status),
  KEY idx_families_searchable (searchable),
  KEY idx_families_public_display_status (public_display_status),
  KEY idx_families_graph_version (graph_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 7. family_members

```sql
CREATE TABLE family_members (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  member_type VARCHAR(40) NOT NULL DEFAULT 'LINEAGE_MEMBER',
  surname VARCHAR(40) NULL,
  generation_character VARCHAR(40) NULL,
  given_name VARCHAR(100) NULL,
  display_name VARCHAR(150) NOT NULL,
  gender VARCHAR(40) NOT NULL DEFAULT 'UNKNOWN',
  birth_date DATE NULL,
  death_date DATE NULL,
  birth_order INT NULL,
  manual_order INT NULL,
  native_place VARCHAR(200) NULL,
  region_text VARCHAR(200) NULL,
  is_living TINYINT NULL,
  user_binding_policy VARCHAR(40) NOT NULL DEFAULT 'OPTIONAL',
  unbound_reason VARCHAR(80) NULL,
  unbound_note VARCHAR(500) NULL,
  manual_lineage_override TINYINT NOT NULL DEFAULT 0,
  lineage_note_type VARCHAR(80) NULL,
  lineage_note VARCHAR(500) NULL,
  is_terminal_node TINYINT NOT NULL DEFAULT 0,
  terminal_reason VARCHAR(100) NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  created_by_user_id BIGINT NULL,
  created_by_admin_id BIGINT NULL,
  deleted_at DATETIME NULL,
  deleted_by_user_id BIGINT NULL,
  deleted_by_admin_id BIGINT NULL,
  delete_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fm_family_id (family_id),
  KEY idx_fm_display_name (display_name),
  KEY idx_fm_status (status),
  KEY idx_fm_member_type (member_type),
  KEY idx_fm_user_binding_policy (user_binding_policy)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 8. family_relationships

```sql
CREATE TABLE family_relationships (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  from_member_id BIGINT NOT NULL,
  to_member_id BIGINT NOT NULL,
  relationship_type VARCHAR(40) NOT NULL,
  parent_link_type VARCHAR(40) NULL,
  relation_note_type VARCHAR(80) NULL,
  relation_note VARCHAR(500) NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  created_by_user_id BIGINT NULL,
  created_by_admin_id BIGINT NULL,
  deleted_at DATETIME NULL,
  deleted_by_user_id BIGINT NULL,
  deleted_by_admin_id BIGINT NULL,
  delete_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fr_family_id (family_id),
  KEY idx_fr_from_member_id (from_member_id),
  KEY idx_fr_to_member_id (to_member_id),
  KEY idx_fr_relationship_type (relationship_type),
  KEY idx_fr_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

业务规则：relationship_type 只允许 PARENT_CHILD / SPOUSE；不保存 SIBLING。

## 9. family_member_user_links

```sql
CREATE TABLE family_member_user_links (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  member_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  link_status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  link_source VARCHAR(80) NOT NULL,
  family_role VARCHAR(40) NOT NULL DEFAULT 'MEMBER',
  invitation_id BIGINT NULL,
  join_request_id BIGINT NULL,
  role_granted_at DATETIME NULL,
  role_granted_by_user_id BIGINT NULL,
  role_granted_by_admin_id BIGINT NULL,
  unlinked_at DATETIME NULL,
  unlinked_reason VARCHAR(500) NULL,
  unlinked_by_user_id BIGINT NULL,
  unlinked_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fmul_family_id (family_id),
  KEY idx_fmul_member_id (member_id),
  KEY idx_fmul_user_id (user_id),
  KEY idx_fmul_link_status (link_status),
  KEY idx_fmul_family_role (family_role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

Service 层强校验：一个 user 在同一 family 只能有一个 ACTIVE link；一个 member 只能有一个 ACTIVE user link；同一 family 只能有一个 FOUNDER。

## 10. invitations / join requests / public applications / messages

```sql
CREATE TABLE family_invitations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  target_member_id BIGINT NOT NULL,
  inviter_user_id BIGINT NULL,
  inviter_admin_id BIGINT NULL,
  target_user_id BIGINT NULL,
  target_phone VARCHAR(30) NULL,
  target_phone_hash VARCHAR(128) NULL,
  invite_type VARCHAR(80) NOT NULL,
  invite_channel VARCHAR(40) NOT NULL,
  invite_actor_type VARCHAR(80) NOT NULL,
  family_role_after_accept VARCHAR(40) NOT NULL DEFAULT 'MEMBER',
  invite_token VARCHAR(180) NULL,
  invite_message VARCHAR(500) NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  accepted_by_user_id BIGINT NULL,
  accepted_at DATETIME NULL,
  rejected_by_user_id BIGINT NULL,
  rejected_at DATETIME NULL,
  reject_reason VARCHAR(500) NULL,
  cancelled_by_user_id BIGINT NULL,
  cancelled_by_admin_id BIGINT NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  expired_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fi_family_id (family_id),
  KEY idx_fi_target_member_id (target_member_id),
  KEY idx_fi_invite_token (invite_token),
  KEY idx_fi_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_join_requests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  applicant_user_id BIGINT NOT NULL,
  applicant_real_name VARCHAR(100) NULL,
  applicant_phone VARCHAR(30) NULL,
  applicant_phone_hash VARCHAR(128) NULL,
  applicant_message VARCHAR(500) NULL,
  request_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  approve_mode VARCHAR(80) NULL,
  bound_member_id BIGINT NULL,
  created_member_id BIGINT NULL,
  handle_result VARCHAR(80) NULL,
  handled_by_user_id BIGINT NULL,
  handled_by_admin_id BIGINT NULL,
  handled_at DATETIME NULL,
  handle_comment VARCHAR(500) NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fjr_family_id (family_id),
  KEY idx_fjr_applicant_user_id (applicant_user_id),
  KEY idx_fjr_request_status (request_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_public_applications (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  applicant_user_id BIGINT NULL,
  applicant_admin_id BIGINT NULL,
  application_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  application_reason VARCHAR(500) NULL,
  application_snapshot_json JSON NULL,
  review_result VARCHAR(40) NULL,
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fpa_family_id (family_id),
  KEY idx_fpa_status (application_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE visitor_messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  visitor_user_id BIGINT NULL,
  visitor_name VARCHAR(100) NULL,
  visitor_phone VARCHAR(30) NULL,
  visitor_wechat VARCHAR(120) NULL,
  message_content VARCHAR(1000) NOT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  deleted_at DATETIME NULL,
  deleted_by_admin_id BIGINT NULL,
  delete_reason VARCHAR(500) NULL,
  ip VARCHAR(80) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_vm_family_id (family_id),
  KEY idx_vm_status (status),
  KEY idx_vm_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 11. admin_users

```sql
CREATE TABLE admin_users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(100) NULL,
  phone VARCHAR(30) NULL,
  email VARCHAR(150) NULL,
  role VARCHAR(40) NOT NULL,
  status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE',
  failed_login_count INT NOT NULL DEFAULT 0,
  locked_until DATETIME NULL,
  last_login_at DATETIME NULL,
  last_login_ip VARCHAR(80) NULL,
  password_changed_at DATETIME NULL,
  disabled_at DATETIME NULL,
  disabled_by_admin_id BIGINT NULL,
  disabled_reason VARCHAR(500) NULL,
  deleted_at DATETIME NULL,
  deleted_by_admin_id BIGINT NULL,
  unlocked_at DATETIME NULL,
  unlocked_by_admin_id BIGINT NULL,
  remark VARCHAR(500) NULL,
  created_by_admin_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_admin_users_username (username),
  KEY idx_admin_users_role (role),
  KEY idx_admin_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 12. transfer / dissolution

```sql
CREATE TABLE family_founder_transfer_requests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  from_member_id BIGINT NOT NULL,
  from_user_id BIGINT NOT NULL,
  to_member_id BIGINT NOT NULL,
  to_user_id BIGINT NOT NULL,
  request_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  request_reason VARCHAR(500) NULL,
  review_result VARCHAR(40) NULL,
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  completed_at DATETIME NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fftr_family_id (family_id),
  KEY idx_fftr_status (request_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_dissolution_requests (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  requester_member_id BIGINT NOT NULL,
  requester_user_id BIGINT NOT NULL,
  request_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
  request_reason VARCHAR(500) NULL,
  request_snapshot_json JSON NULL,
  review_result VARCHAR(40) NULL,
  reviewed_by_admin_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  review_comment VARCHAR(500) NULL,
  completed_at DATETIME NULL,
  cancelled_at DATETIME NULL,
  cancel_reason VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fdr_family_id (family_id),
  KEY idx_fdr_status (request_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 13. operation_logs

```sql
CREATE TABLE operation_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  operator_type VARCHAR(40) NOT NULL,
  operator_admin_id BIGINT NULL,
  operator_user_id BIGINT NULL,
  operator_role VARCHAR(60) NULL,
  module VARCHAR(80) NOT NULL,
  action VARCHAR(100) NOT NULL,
  target_type VARCHAR(80) NULL,
  target_id BIGINT NULL,
  family_id BIGINT NULL,
  member_id BIGINT NULL,
  user_id BIGINT NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  detail_json JSON NULL,
  result VARCHAR(30) NOT NULL DEFAULT 'SUCCESS',
  error_message VARCHAR(500) NULL,
  ip VARCHAR(80) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_operation_logs_operator_admin (operator_admin_id),
  KEY idx_operation_logs_operator_user (operator_user_id),
  KEY idx_operation_logs_module_action (module, action),
  KEY idx_operation_logs_target (target_type, target_id),
  KEY idx_operation_logs_family (family_id),
  KEY idx_operation_logs_member (member_id),
  KEY idx_operation_logs_user (user_id),
  KEY idx_operation_logs_result (result),
  KEY idx_operation_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 14. 必须由 Service 层强校验的规则

```text
1. 一个 user 在同一 family 只能有一个 ACTIVE member link。
2. 一个 member 只能有一个 ACTIVE user link。
3. 同一 family 只能有一个 FOUNDER。
4. 同一 child 的 PRIMARY 父亲最多一个。
5. 同一 child 的 PRIMARY 母亲最多一个。
6. 同一 member 同时只能有一个 PENDING invitation。
7. 同一 family 同时只能有一个 PENDING public application。
8. 同一 family 同时只能有一个 PENDING founder transfer request。
9. 同一 family 同时只能有一个 PENDING dissolution request。
10. ROOT_ADMIN 有且只有一个。
```

## 15. Migration 拆分建议

```text
000001_create_users.up.sql
000002_create_user_auth_identities.up.sql
000003_create_account_logs.up.sql
000004_create_families.up.sql
000005_create_family_members.up.sql
000006_create_family_relationships.up.sql
000007_create_family_member_user_links.up.sql
000008_create_family_invitations.up.sql
000009_create_family_join_requests.up.sql
000010_create_public_applications_and_messages.up.sql
000011_create_admin_users.up.sql
000012_create_transfer_and_dissolution.up.sql
000013_create_operation_logs.up.sql
000014_seed_root_admin.up.sql
```
