package service

import (
	"crypto/sha512"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	midtranssdk "github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"

	"pulsa2/internal/helper"
	"pulsa2/internal/repository"
)

type AppOrderPaymentService struct {
	paymentRepo    *repository.AppOrderPaymentRepository
	orderRepo      *repository.AppOrderRepository
	bankRepo       *repository.BankRepository
	fulfillmentSvc *AppOrderFulfillmentService
}

func NewAppOrderPaymentService(paymentRepo *repository.AppOrderPaymentRepository, orderRepo *repository.AppOrderRepository, bankRepo *repository.BankRepository, fulfillmentSvc *AppOrderFulfillmentService) *AppOrderPaymentService {
	return &AppOrderPaymentService{paymentRepo: paymentRepo, orderRepo: orderRepo, bankRepo: bankRepo, fulfillmentSvc: fulfillmentSvc}
}
func verifyMidtransSignature(orderID, statusCode, grossAmount, serverKey, signatureKey string) bool {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	return strings.EqualFold(fmt.Sprintf("%x", sum[:]), strings.TrimSpace(signatureKey))
}

func isMidtransPaid(transactionStatus, fraudStatus, statusCode string) bool {
	transactionStatus = strings.TrimSpace(strings.ToLower(transactionStatus))
	fraudStatus = strings.TrimSpace(strings.ToLower(fraudStatus))
	statusCode = strings.TrimSpace(statusCode)
	if statusCode != "200" {
		return false
	}
	if fraudStatus != "" && fraudStatus != "accept" {
		return false
	}
	return transactionStatus == "settlement" || transactionStatus == "capture"
}

func mapOrderStatusFromMidtrans(transactionStatus, fraudStatus, statusCode string) string {
	transactionStatus = strings.TrimSpace(strings.ToLower(transactionStatus))
	if isMidtransPaid(transactionStatus, fraudStatus, statusCode) {
		return "paid"
	}
	switch transactionStatus {
	case "pending":
		return "pending_payment"
	case "expire":
		return "expired"
	case "cancel":
		return "cancelled"
	case "deny", "failure":
		return "failed"
	default:
		return "pending_payment"
	}
}

func firstQRURLFromPayload(payload map[string]any) string {
	actions, ok := payload["actions"].([]any)
	if !ok {
		return ""
	}
	for _, action := range actions {
		m, ok := action.(map[string]any)
		if !ok {
			continue
		}
		url := strings.TrimSpace(helper.ToString(m["url"]))
		if url != "" {
			return url
		}
	}
	return ""
}

func firstQRURL(actions []coreapi.Action) string {
	for _, action := range actions {
		if url := strings.TrimSpace(action.URL); url != "" {
			return url
		}
	}
	return ""
}

func buildMidtransChargeRequest(order *repository.AppOrderRow, qrisAmount int64) *coreapi.ChargeReq {
	items := []midtranssdk.ItemDetails{{
		ID:    order.ProdukSKUSnapshot,
		Price: qrisAmount,
		Qty:   1,
		Name:  trimMidtransItemName(order.ProdukNamaSnapshot),
	}}
	req := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeQris,
		TransactionDetails: midtranssdk.TransactionDetails{
			OrderID:  order.InvoiceID,
			GrossAmt: qrisAmount,
		},
		Items: &items,
		Qris: &coreapi.QrisDetails{
			Acquirer: midtransAcquirer(),
		},
	}
	if strings.EqualFold(strings.TrimSpace(order.BuyerType), "guest") {
		req.CustomerDetails = &midtranssdk.CustomerDetails{
			FName: derefString(order.GuestNama),
			Email: derefString(order.GuestEmail),
			Phone: derefString(order.GuestPhone),
		}
	}
	return req
}

func chargeMidtransQRIS(serverKey string, payload *coreapi.ChargeReq) (*coreapi.ChargeResponse, error) {
	client := newMidtransCoreAPIClient(serverKey)
	resp, err := client.ChargeTransaction(payload)
	if !isNilLikeError(err) {
		msg := strings.TrimSpace(err.Message)
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("midtrans charge gagal: %s", msg)
	}
	return resp, nil
}

func isNilLikeError(err error) bool {
	if err == nil {
		return true
	}
	v := reflect.ValueOf(err)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func parseMidtransTime(v string) *time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
	if err != nil {
		return nil
	}
	return &t
}

func newMidtransCoreAPIClient(serverKey string) *coreapi.Client {
	env := midtranssdk.Sandbox
	if isMidtransProduction() {
		env = midtranssdk.Production
	}
	client := &coreapi.Client{}
	client.New(serverKey, env)
	return client
}

func midtransServerKeyFromEnv() (string, error) {
	serverKey := strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY"))
	if err := validateMidtransServerKey(serverKey, isMidtransProduction()); err != nil {
		return "", err
	}
	return serverKey, nil
}

func validateMidtransServerKey(serverKey string, production bool) error {
	serverKey = strings.TrimSpace(serverKey)
	if serverKey == "" {
		return fmt.Errorf("MIDTRANS_SERVER_KEY belum diset")
	}

	lowerKey := strings.ToLower(serverKey)
	for _, marker := range []string{"ganti", "change_me", "your_", "your-", "placeholder", "dummy"} {
		if strings.Contains(lowerKey, marker) {
			return fmt.Errorf("MIDTRANS_SERVER_KEY masih placeholder")
		}
	}

	if production {
		if strings.HasPrefix(serverKey, "SB-Mid-server-") {
			return fmt.Errorf("MIDTRANS_IS_PRODUCTION=true tapi MIDTRANS_SERVER_KEY masih sandbox")
		}
		if !strings.HasPrefix(serverKey, "Mid-server-") {
			return fmt.Errorf("format MIDTRANS_SERVER_KEY production tidak valid; gunakan awalan Mid-server yang valid")
		}
		return nil
	}

	if !strings.HasPrefix(serverKey, "SB-Mid-server-") && !strings.HasPrefix(serverKey, "Mid-server-") {
		return fmt.Errorf("format MIDTRANS_SERVER_KEY tidak valid; gunakan Server Key dari Settings > Access Keys")
	}
	return nil
}

func isMidtransProduction() bool {
	return isTrue(os.Getenv("MIDTRANS_IS_PRODUCTION"))
}

func midtransAcquirer() string {
	if v := strings.TrimSpace(os.Getenv("MIDTRANS_QRIS_ACQUIRER")); v != "" {
		return strings.ToLower(v)
	}
	return "gopay"
}

func trimMidtransItemName(v string) string {
	v = strings.TrimSpace(v)
	if len(v) <= 50 {
		return v
	}
	return strings.TrimSpace(v[:50])
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func isTrue(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "1" || v == "true" || v == "yes" || v == "y"
}
