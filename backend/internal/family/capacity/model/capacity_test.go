package model

import "testing"

func TestCapacityStateConstantsRemainExplicit(t *testing.T) {
	if ProductStatusActive == "" || OrderStatusPaid == "" || EntitlementActive == "" || RefundStatusSuccess == "" {
		t.Fatal("capacity state constants must be explicit database values")
	}
	if OrderStatusPaid == OrderStatusRefunded || EntitlementActive == EntitlementRevoked {
		t.Fatal("capacity state constants must be distinguishable")
	}
}
