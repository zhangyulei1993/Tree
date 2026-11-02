package wechatpay

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"tree/backend/internal/common/config"
	capacitydto "tree/backend/internal/family/capacity/dto"
	capacitymodel "tree/backend/internal/family/capacity/model"
)

type Provider struct {
	client       *core.Client
	cfg          config.Config
	notify       *notify.Handler
	orderTimeout time.Duration
}

func New(ctx context.Context, cfg config.Config) (*Provider, error) {
	if !cfg.WechatPay.Enabled {
		return nil, fmt.Errorf("wechat pay is disabled")
	}
	key, err := utils.LoadPrivateKeyWithPath(cfg.WechatPay.MerchantPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load merchant private key: %w", err)
	}
	var client *core.Client
	var verifier auth.Verifier
	if cfg.WechatPay.PublicKeyID != "" && cfg.WechatPay.PublicKeyPath != "" {
		publicKey, loadErr := utils.LoadPublicKeyWithPath(cfg.WechatPay.PublicKeyPath)
		if loadErr != nil {
			return nil, fmt.Errorf("load wechat pay public key: %w", loadErr)
		}
		client, err = core.NewClient(ctx, option.WithWechatPayPublicKeyAuthCipher(cfg.WechatPay.MerchantID, cfg.WechatPay.MerchantCertificateSerialNo, key, cfg.WechatPay.PublicKeyID, publicKey))
		verifier = verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.WechatPay.PublicKeyID, *publicKey)
	} else {
		client, err = core.NewClient(ctx, option.WithWechatPayAutoAuthCipher(cfg.WechatPay.MerchantID, cfg.WechatPay.MerchantCertificateSerialNo, key, cfg.WechatPay.APIv3Key))
		verifier = verifiers.NewSHA256WithRSAVerifier(downloader.MgrInstance().GetCertificateVisitor(cfg.WechatPay.MerchantID))
	}
	if err != nil {
		return nil, fmt.Errorf("create wechat pay client: %w", err)
	}
	notifyHandler, err := notify.NewRSANotifyHandler(cfg.WechatPay.APIv3Key, verifier)
	if err != nil {
		return nil, fmt.Errorf("create wechat notify handler: %w", err)
	}
	orderTimeout := time.Duration(cfg.Capacity.OrderTimeoutMinutes) * time.Minute
	if orderTimeout <= 0 {
		orderTimeout = 30 * time.Minute
	}
	return &Provider{client: client, cfg: cfg, notify: notifyHandler, orderTimeout: orderTimeout}, nil
}

type PaymentNotification struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	AppID         string `json:"appid"`
	MchID         string `json:"mchid"`
	Amount        struct {
		Total      int64 `json:"total"`
		PayerTotal int64 `json:"payer_total"`
	} `json:"amount"`
}
type RefundNotification struct {
	OutRefundNo   string `json:"out_refund_no"`
	RefundID      string `json:"refund_id"`
	RefundStatus  string `json:"refund_status"`
	TransactionID string `json:"transaction_id"`
	Amount        struct {
		Total  int64 `json:"total"`
		Refund int64 `json:"refund"`
	} `json:"amount"`
}

func (p *Provider) ParsePaymentNotification(ctx context.Context, req *http.Request) (*PaymentNotification, error) {
	var content PaymentNotification
	if _, err := p.notify.ParseNotifyRequest(ctx, req, &content); err != nil {
		return nil, err
	}
	return &content, nil
}
func (p *Provider) ParseRefundNotification(ctx context.Context, req *http.Request) (*RefundNotification, error) {
	var content RefundNotification
	if _, err := p.notify.ParseNotifyRequest(ctx, req, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

func (p *Provider) CreateJSAPI(ctx context.Context, order *capacitymodel.Order, openID string) (capacitydto.PaymentParams, error) {
	if p == nil || p.client == nil || openID == "" {
		return capacitydto.PaymentParams{}, fmt.Errorf("wechat pay provider is not ready")
	}
	total := int64(order.AmountFen)
	svc := jsapi.JsapiApiService{Client: p.client}
	resp, _, err := svc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(p.cfg.Wechat.MiniAppID),
		Mchid:       core.String(p.cfg.WechatPay.MerchantID),
		Description: core.String(order.ProductName),
		OutTradeNo:  core.String(order.OrderNo),
		TimeExpire:  core.Time(time.Now().Add(p.orderTimeout)),
		NotifyUrl:   core.String(p.cfg.WechatPay.NotifyURL),
		Amount:      &jsapi.Amount{Total: &total, Currency: core.String("CNY")},
		Payer:       &jsapi.Payer{Openid: core.String(openID)},
	})
	if err != nil {
		return capacitydto.PaymentParams{}, err
	}
	return capacitydto.PaymentParams{TimeStamp: deref(resp.TimeStamp), NonceStr: deref(resp.NonceStr), Package: deref(resp.Package), SignType: deref(resp.SignType), PaySign: deref(resp.PaySign)}, nil
}

func (p *Provider) QueryPayment(ctx context.Context, orderNo string) (capacitydto.PaymentQueryResult, error) {
	if p == nil || p.client == nil || orderNo == "" {
		return capacitydto.PaymentQueryResult{}, fmt.Errorf("wechat pay provider is not ready")
	}
	svc := jsapi.JsapiApiService{Client: p.client}
	row, _, err := svc.QueryOrderByOutTradeNo(ctx, jsapi.QueryOrderByOutTradeNoRequest{OutTradeNo: core.String(orderNo), Mchid: core.String(p.cfg.WechatPay.MerchantID)})
	if err != nil {
		var apiErr *core.APIError
		if errors.As(err, &apiErr) && apiErr.Code == "RESOURCE_NOT_EXISTS" {
			return capacitydto.PaymentQueryResult{OrderNo: orderNo, Status: "NOT_FOUND"}, nil
		}
		return capacitydto.PaymentQueryResult{}, err
	}
	if row == nil || row.OutTradeNo == nil || row.TradeState == nil {
		return capacitydto.PaymentQueryResult{}, fmt.Errorf("wechat payment query response is incomplete")
	}
	result := capacitydto.PaymentQueryResult{OrderNo: *row.OutTradeNo, Status: string(*row.TradeState), AppID: deref(row.Appid), MerchantID: deref(row.Mchid)}
	if row.TransactionId != nil {
		result.TransactionID = *row.TransactionId
	}
	if row.Amount != nil {
		if row.Amount.Total != nil {
			result.TotalAmount = *row.Amount.Total
		}
		if row.Amount.PayerTotal != nil {
			result.PayerTotal = *row.Amount.PayerTotal
		}
	}
	return result, nil
}

func (p *Provider) Refund(ctx context.Context, order *capacitymodel.Order, refund *capacitymodel.Refund) error {
	if p == nil || p.client == nil || order.WechatTransactionID == nil {
		return fmt.Errorf("wechat payment transaction is missing")
	}
	total := int64(order.AmountFen)
	svc := refunddomestic.RefundsApiService{Client: p.client}
	_, _, err := svc.Create(ctx, refunddomestic.CreateRequest{
		TransactionId: order.WechatTransactionID,
		OutTradeNo:    core.String(order.OrderNo), OutRefundNo: core.String(refund.RefundNo),
		Reason: core.String("家庭容量扩容退款"), NotifyUrl: core.String(p.cfg.WechatPay.RefundNotifyURL),
		Amount: &refunddomestic.AmountReq{Refund: &total, Total: &total, Currency: core.String("CNY")},
	})
	return err
}

func (p *Provider) QueryRefund(ctx context.Context, refundNo string) (capacitydto.RefundQueryResult, error) {
	if p == nil || p.client == nil || refundNo == "" {
		return capacitydto.RefundQueryResult{}, fmt.Errorf("wechat pay provider is not ready")
	}
	svc := refunddomestic.RefundsApiService{Client: p.client}
	row, _, err := svc.QueryByOutRefundNo(ctx, refunddomestic.QueryByOutRefundNoRequest{OutRefundNo: core.String(refundNo)})
	if err != nil {
		var apiErr *core.APIError
		if errors.As(err, &apiErr) && apiErr.Code == "RESOURCE_NOT_EXISTS" {
			return capacitydto.RefundQueryResult{RefundNo: refundNo, Status: "NOT_FOUND"}, nil
		}
		return capacitydto.RefundQueryResult{}, err
	}
	if row == nil || row.RefundId == nil || row.OutRefundNo == nil || row.OutTradeNo == nil || row.TransactionId == nil || row.Status == nil || row.Amount == nil || row.Amount.Total == nil || row.Amount.Refund == nil {
		return capacitydto.RefundQueryResult{}, fmt.Errorf("wechat refund query response is incomplete")
	}
	return capacitydto.RefundQueryResult{
		RefundNo: *row.OutRefundNo, RefundID: *row.RefundId,
		OrderNo: *row.OutTradeNo, TransactionID: *row.TransactionId,
		Status: string(*row.Status), TotalAmount: *row.Amount.Total, RefundAmount: *row.Amount.Refund,
	}, nil
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
