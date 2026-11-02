-- Read-path indexes for the family tree, invitations, links and capacity orders.
-- These are additive and preserve existing data and business semantics.
ALTER TABLE family_members
  ADD KEY idx_fm_family_status_deleted_order (family_id, status, deleted_at, manual_order, id);

ALTER TABLE family_relationships
  ADD KEY idx_fr_family_status_deleted_endpoints (family_id, status, deleted_at, from_member_id, to_member_id),
  ADD KEY idx_fr_family_status_deleted_type (family_id, status, deleted_at, relationship_type);

ALTER TABLE family_member_user_links
  ADD KEY idx_fmul_family_member_status_user (family_id, member_id, link_status, user_id),
  ADD KEY idx_fmul_family_user_status_member (family_id, user_id, link_status, member_id);

ALTER TABLE family_invitations
  ADD KEY idx_fi_family_member_status_expiry (family_id, target_member_id, status, expired_at, id),
  ADD KEY idx_fi_family_status_id (family_id, status, id),
  ADD KEY idx_fi_target_user_status_id (target_user_id, status, id);

ALTER TABLE family_capacity_orders
  ADD KEY idx_capacity_orders_family_buyer_id (family_id, buyer_user_id, id);
