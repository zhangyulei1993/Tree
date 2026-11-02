ALTER TABLE family_capacity_orders
  DROP KEY idx_capacity_orders_family_buyer_id;

ALTER TABLE family_invitations
  DROP KEY idx_fi_family_member_status_expiry,
  DROP KEY idx_fi_family_status_id,
  DROP KEY idx_fi_target_user_status_id;

ALTER TABLE family_member_user_links
  DROP KEY idx_fmul_family_member_status_user,
  DROP KEY idx_fmul_family_user_status_member;

ALTER TABLE family_relationships
  DROP KEY idx_fr_family_status_deleted_endpoints,
  DROP KEY idx_fr_family_status_deleted_type;

ALTER TABLE family_members
  DROP KEY idx_fm_family_status_deleted_order;
