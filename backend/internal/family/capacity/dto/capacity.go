package dto

import "time"

type CreateOrderRequest struct {
	ProductKey string `json:"productKey" binding:"required"`
}

type UpdateProductRequest struct {
	DisplayName         string `json:"displayName"`
	AddedMembers        int    `json:"addedMembers"`
	AmountFen           int    `json:"amountFen"`
	RefundWindowMinutes int    `json:"refundWindowMinutes"`
	Status              string `json:"status"`
}

type OrderVO struct {
	OrderNo                 string         `json:"orderNo"`
	FamilyID                uint64         `json:"familyId"`
	ProductName             string         `json:"productName"`
	AddedMembers            int            `json:"addedMembers"`
	AmountFen               int            `json:"amountFen"`
	Status                  string         `json:"status"`
	RefundWindowMinutes     int            `json:"refundWindowMinutes"`
	RefundableUntil         string         `json:"refundableUntil,omitempty"`
	RefundAttemptCount      int            `json:"refundAttemptCount"`
	RefundAttemptsRemaining int            `json:"refundAttemptsRemaining"`
	LastRefundStatus        string         `json:"lastRefundStatus,omitempty"`
	LastRefundReason        string         `json:"lastRefundReason,omitempty"`
	ExpiresAt               string         `json:"expiresAt,omitempty"`
	ClosedAt                string         `json:"closedAt,omitempty"`
	Payment                 *PaymentParams `json:"payment,omitempty"`
	CreatedAt               string         `json:"createdAt"`
}

type ProductVO struct {
	ProductKey          string `json:"productKey"`
	DisplayName         string `json:"displayName"`
	AddedMembers        int    `json:"addedMembers"`
	AmountFen           int    `json:"amountFen"`
	RefundWindowMinutes int    `json:"refundWindowMinutes"`
	Status              string `json:"status,omitempty"`
}

type AdminOrderVO struct {
	OrderNo                 string `json:"orderNo"`
	FamilyID                uint64 `json:"familyId"`
	BuyerUserID             uint64 `json:"buyerUserId"`
	ProductName             string `json:"productName"`
	AddedMembers            int    `json:"addedMembers"`
	AmountFen               int    `json:"amountFen"`
	Status                  string `json:"status"`
	TransactionID           string `json:"transactionId,omitempty"`
	CreatedAt               string `json:"createdAt"`
	ExpiresAt               string `json:"expiresAt,omitempty"`
	PaidAt                  string `json:"paidAt,omitempty"`
	ClosedAt                string `json:"closedAt,omitempty"`
	RefundNo                string `json:"refundNo,omitempty"`
	RefundStatus            string `json:"refundStatus,omitempty"`
	RefundRequestedAt       string `json:"refundRequestedAt,omitempty"`
	RefundReason            string `json:"refundReason,omitempty"`
	RefundAttemptNo         int    `json:"refundAttemptNo,omitempty"`
	RefundAttemptsRemaining int    `json:"refundAttemptsRemaining"`
}

type AdminRefundRequest struct {
	Reason string `json:"reason"`
}

type RefundQueryResult struct {
	RefundNo      string
	RefundID      string
	OrderNo       string
	TransactionID string
	Status        string
	TotalAmount   int64
	RefundAmount  int64
}

// PaymentQueryResult is the subset of WeChat's order query response required
// to safely reconcile a capacity order when the asynchronous callback is late.
type PaymentQueryResult struct {
	OrderNo       string
	TransactionID string
	Status        string
	AppID         string
	MerchantID    string
	TotalAmount   int64
	PayerTotal    int64
}

type AdminOrderQuery struct {
	Page      int
	PageSize  int
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
}

type AdminOrderPage struct {
	Items    []AdminOrderVO `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type OrderPage struct {
	Items    []*OrderVO `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

type PaymentParams struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}
