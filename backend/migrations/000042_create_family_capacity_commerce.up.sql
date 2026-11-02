CREATE TABLE family_capacity_products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  product_key VARCHAR(64) NOT NULL,
  display_name VARCHAR(80) NOT NULL,
  added_members INT NOT NULL,
  amount_fen INT NOT NULL,
  refund_window_minutes INT NOT NULL DEFAULT 30,
  status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_family_capacity_products_key (product_key),
  KEY idx_family_capacity_products_status_sort (status, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_capacity_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  family_id BIGINT NOT NULL,
  buyer_user_id BIGINT NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  product_name VARCHAR(80) NOT NULL,
  added_members INT NOT NULL,
  amount_fen INT NOT NULL,
  status VARCHAR(24) NOT NULL,
  wechat_transaction_id VARCHAR(96) NULL,
  paid_at DATETIME NULL,
  closed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_family_capacity_orders_no (order_no),
  UNIQUE KEY uk_family_capacity_orders_transaction (wechat_transaction_id),
  KEY idx_family_capacity_orders_family_status (family_id, status, created_at),
  KEY idx_family_capacity_orders_buyer (buyer_user_id, created_at),
  CONSTRAINT fk_family_capacity_orders_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_capacity_orders_buyer FOREIGN KEY (buyer_user_id) REFERENCES users(id),
  CONSTRAINT fk_family_capacity_orders_product FOREIGN KEY (product_id) REFERENCES family_capacity_products(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_capacity_entitlements (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  added_members INT NOT NULL,
  status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE',
  revoked_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_family_capacity_entitlements_order (order_id),
  KEY idx_family_capacity_entitlements_family_status (family_id, status),
  CONSTRAINT fk_family_capacity_entitlements_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_capacity_entitlements_order FOREIGN KEY (order_id) REFERENCES family_capacity_orders(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE family_capacity_refunds (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  refund_no VARCHAR(64) NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  family_id BIGINT NOT NULL,
  operator_user_id BIGINT NOT NULL,
  amount_fen INT NOT NULL,
  status VARCHAR(24) NOT NULL,
  wechat_refund_id VARCHAR(96) NULL,
  requested_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  succeeded_at DATETIME NULL,
  failed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_family_capacity_refunds_no (refund_no),
  UNIQUE KEY uk_family_capacity_refunds_order (order_id),
  UNIQUE KEY uk_family_capacity_refunds_wechat (wechat_refund_id),
  KEY idx_family_capacity_refunds_family_status (family_id, status, created_at),
  CONSTRAINT fk_family_capacity_refunds_order FOREIGN KEY (order_id) REFERENCES family_capacity_orders(id),
  CONSTRAINT fk_family_capacity_refunds_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_capacity_refunds_operator FOREIGN KEY (operator_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO family_capacity_products (product_key, display_name, added_members, amount_fen, refund_window_minutes, status, sort_order)
VALUES
  ('CAPACITY_PLUS_1', '+1 位成员', 1, 1, 30, 'ACTIVE', 10),
  ('CAPACITY_PLUS_5', '+5 位成员', 5, 1, 30, 'ACTIVE', 20),
  ('CAPACITY_PLUS_10', '+10 位成员', 10, 1, 30, 'ACTIVE', 30),
  ('CAPACITY_PLUS_20', '+20 位成员', 20, 1, 30, 'ACTIVE', 40),
  ('CAPACITY_LARGE_FAMILY', '大家庭扩容 +100 位', 100, 1, 30, 'ACTIVE', 50);
