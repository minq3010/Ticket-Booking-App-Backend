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
	PartnerCode  string `json:"partnerCode"`
	AccessKey    string `json:"accessKey"`
	RequestID    string `json:"requestId"`
	Amount       string `json:"amount"`
	OrderID      string `json:"orderId"`
	OrderInfo    string `json:"orderInfo"`
	RedirectUrl  string `json:"redirectUrl"`
	IpnUrl       string `json:"ipnUrl"`
	ExtraData    string `json:"extraData"`
	RequestType  string `json:"requestType"`
	Signature    string `json:"signature"`
	Lang         string `json:"lang"`
	AutoCapture  bool   `json:"autoCapture"`
}

type MomoPaymentResponse struct {
	PayUrl string `json:"payUrl"`
	ErrorCode int `json:"errorCode"`
	Message string `json:"message"`
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
	redirectURL := os.Getenv("MOMO_REDIRECT_URL")
	ipnURL := os.Getenv("MOMO_IPN_URL")
	orderInfo := "Thanh toán đơn hàng " + orderID
	requestID := fmt.Sprintf("%s-%d", orderID, time.Now().Unix())
	extraData := ""
	requestType := "captureWallet"

	params := map[string]string{
		"accessKey": accessKey,
		"amount": fmt.Sprintf("%d", amount),
		"extraData": extraData,
		"ipnUrl": ipnURL,
		"orderId": orderID,
		"orderInfo": orderInfo,
		"partnerCode": partnerCode,
		"redirectUrl": redirectURL,
		"requestId": requestID,
		"requestType": requestType,
	}

	signature := GenerateMomoSignature(params, secretKey)
	// log Signature
	fmt.Println("Signature: ",signature)
	payload := MomoPaymentRequest{
		PartnerCode: partnerCode,
		AccessKey: accessKey,
		RequestID: requestID,
		Amount: fmt.Sprintf("%d", amount),
		OrderID: orderID,
		OrderInfo: orderInfo,
		RedirectUrl: redirectURL,
		IpnUrl: ipnURL,
		ExtraData: extraData,
		RequestType: requestType,
		Signature: signature,
		Lang: "vi",
		AutoCapture: true,
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var momoResp MomoPaymentResponse
	err = json.NewDecoder(resp.Body).Decode(&momoResp)
	return &momoResp, err
}