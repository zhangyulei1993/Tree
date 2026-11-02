ALTER TABLE family_capacity_orders ADD COLUMN refund_window_minutes INT NOT NULL DEFAULT 30 AFTER amount_fen;
