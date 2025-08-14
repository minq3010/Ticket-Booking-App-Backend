package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type MomoPaymentRequest struct {
	PartnerCode string `json:"partnerCode"`
	AccessKey   string `json:"accessKey"`
	RequestID   string `json:"requestId"`
	Amount      string `json:"amount"`
	OrderID     string `json:"orderId"`
	OrderInfo   string `json:"orderInfo"`
	RedirectUrl string `json:"redirectUrl"`
	IpnUrl      string `json:"ipnUrl"`
	ExtraData   string `json:"extraData"`
	RequestType string `json:"requestType"`
	Signature   string `json:"signature"`
	Lang        string `json:"lang"`
	AutoCapture bool   `json:"autoCapture"`
}

type MomoPaymentResponse struct {
	PayUrl    string `json:"payUrl"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

func GenerateMomoIPNSignature(params map[string]string, secretKey string) string {
	// IPN signature format khác với create payment
	raw := fmt.Sprintf(
		"accessKey=%s&amount=%s&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%s&resultCode=%s&transId=%s",
		params["accessKey"],
		params["amount"],
		params["extraData"],
		params["message"],
		params["orderId"],
		params["orderInfo"],
		params["orderType"],
		params["partnerCode"],
		params["payType"],
		params["requestId"],
		params["responseTime"],
		params["resultCode"],
		params["transId"],
	)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(raw))
	fmt.Println("🔐 IPN Raw signature string:", raw)
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateMomoSignature(params map[string]string, secretKey string) string {
	raw := fmt.Sprintf(
		"accessKey=%s&amount=%s&extraData=%s&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=%s",
		params["accessKey"],
		params["amount"],
		params["extraData"],
		params["ipnUrl"],
		params["orderId"],
		params["orderInfo"],
		params["partnerCode"],
		params["redirectUrl"],
		params["requestId"],
		params["requestType"],
	)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(raw))
	fmt.Println(raw)
	return hex.EncodeToString(h.Sum(nil))
}

func CreateMomoPayment(orderID string, amount int) (*MomoPaymentResponse, error) {
	endpoint := "https://test-payment.momo.vn/v2/gateway/api/create"

	accessKey := os.Getenv("MOMO_ACCESS_KEY")
	secretKey := os.Getenv("MOMO_SECRET_KEY")
	partnerCode := os.Getenv("MOMO_PARTNER_CODE")
	ngrokURL := os.Getenv("NGROK_URL")
    redirectURL := fmt.Sprintf("%s/api/payment-callback/momo-return", ngrokURL)
    ipnURL := fmt.Sprintf("%s/api/payment-callback/momo-ipn", ngrokURL)
	useLocalhost := os.Getenv("USE_LOCALHOST")
	if useLocalhost == "true" {
		redirectURL = "http://localhost:26367/api/payment-callback/momo-return"
		ipnURL = "http://localhost:26367/api/payment-callback/momo-ipn"
		fmt.Println("⚡ Using localhost URLs for MoMo callbacks")
	}
	// ✅ Debug env variables
	fmt.Printf("🔧 MoMo Config Check:\n")
	fmt.Printf("  MOMO_IPN_URL: %s\n", ipnURL)
	fmt.Printf("  MOMO_REDIRECT_URL: %s\n", redirectURL)

	orderInfo := "Thanh toán đơn hàng " + orderID
	requestID := fmt.Sprintf("%s-%d", orderID, time.Now().Unix())
	extraData := ""
	requestType := "captureWallet"

	params := map[string]string{
		"accessKey":   accessKey,
		"amount":      fmt.Sprintf("%d", amount),
		"extraData":   extraData,
		"ipnUrl":      ipnURL,
		"orderId":     orderID,
		"orderInfo":   orderInfo,
		"partnerCode": partnerCode,
		"redirectUrl": redirectURL,
		"requestId":   requestID,
		"requestType": requestType,
	}

	signature := GenerateMomoSignature(params, secretKey)
	fmt.Printf("🔐 Signature: %s\n", signature)

	payload := MomoPaymentRequest{
		PartnerCode: partnerCode,
		AccessKey:   accessKey,
		RequestID:   requestID,
		Amount:      fmt.Sprintf("%d", amount),
		OrderID:     orderID,
		OrderInfo:   orderInfo,
		RedirectUrl: redirectURL,
		IpnUrl:      ipnURL, // ✅ Đảm bảo có IPN URL
		ExtraData:   extraData,
		RequestType: requestType,
		Signature:   signature,
		Lang:        "vi",
		AutoCapture: true,
	}

	// ✅ Debug payload đúng cách
	fmt.Printf("🚀 MoMo Payload:\n")
	fmt.Printf("  OrderID: %s\n", payload.OrderID)
	fmt.Printf("  OrderInfo: %s\n", payload.OrderInfo)
	fmt.Printf("  RedirectUrl: %s\n", payload.RedirectUrl)
	fmt.Printf("  IpnUrl: %s\n", payload.IpnUrl) // ✅ Sửa lại
	fmt.Printf("  Amount: %s\n", payload.Amount)

	jsonPayload, _ := json.Marshal(payload)

	// ✅ Debug raw JSON để kiểm tra
	fmt.Printf("📦 Raw JSON Payload: %s\n", string(jsonPayload))

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var momoResp MomoPaymentResponse
	err = json.NewDecoder(resp.Body).Decode(&momoResp)

	// ✅ Debug MoMo response
	fmt.Printf("📬 MoMo Response:\n")
	fmt.Printf("  ErrorCode: %d\n", momoResp.ErrorCode)
	fmt.Printf("  Message: %s\n", momoResp.Message)
	fmt.Printf("  PayUrl: %s\n", momoResp.PayUrl)

	return &momoResp, err
}

func GenerateOrderID(userID uint, eventID uint) string {
	return fmt.Sprintf("ORDER_%d_%d_%d", userID, eventID, time.Now().Unix())
}
