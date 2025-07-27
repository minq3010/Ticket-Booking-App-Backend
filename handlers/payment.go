package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/minq3010/Backend-React-Native-App/models"
	"github.com/minq3010/Backend-React-Native-App/utils"
)

type PaymentHandler struct {
	PaymentRepo models.PaymentRepository
	EventRepo   models.EventRepository
	TicketRepo  models.TicketRepository
}

// POST /payment/vnpay
func (h *PaymentHandler) CreateVnpayCheckout(c *fiber.Ctx) error {
	type Body struct {
		EventID uint `json:"eventId"`
	}
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	userId := uint(c.Locals("userId").(float64))
	ctx := context.Background()

	event, err := h.EventRepo.GetOne(ctx, body.EventID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "event not found",
		})
	}

	orderID := utils.GenerateOrderID(userId, event.ID)

	payment := &models.Payment{
		OrderID: orderID,
		EventID: event.ID,
		UserID:  userId,
		Amount:  int(event.Price), 
		Status:  "pending",
		Method:  "vnpay",
	}

	if err := h.PaymentRepo.Create(ctx, payment); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create payment",
		})
	}

	payURL, err := utils.CreateVnpayURL(orderID, payment.Amount)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate payment URL",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"url":    payURL,
	})
}

// GET /payment/vnpay-return
func (h *PaymentHandler) HandleVnpayReturn(c *fiber.Ctx) error {
	// 🔧 Cách 1: Sử dụng c.Query() để lấy từng parameter cụ thể
	orderID := c.Query("vnp_TxnRef")
	responseCode := c.Query("vnp_ResponseCode")
	signature := c.Query("vnp_SecureHash")

	// 🔧 Cách 2: Hoặc tạo map từ query string thủ công
	queryParams := make(map[string]string)

	// Lấy tất cả query parameters
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		queryParams[string(key)] = string(value)
	})

	// 👉 Debug log để kiểm tra
	fmt.Println("🔍 OrderID:", orderID)
	fmt.Println("🔍 ResponseCode:", responseCode)
	fmt.Println("🔍 Signature:", signature)
	fmt.Println("🔍 All query params:", queryParams)
	fmt.Println("🔍 Raw query string:", string(c.Context().QueryArgs().QueryString()))

	// 1. Kiểm tra các tham số bắt buộc
	if orderID == "" || responseCode == "" || signature == "" {
		return c.Status(http.StatusBadRequest).SendString("❌ Thiếu thông tin callback từ VNPay")
	}

	// 2. Xác thực chữ ký
	// if !utils.VerifyVnpaySignature(queryParams, signature) {
	// 	return c.Status(http.StatusBadRequest).SendString("❌ Sai chữ ký, không hợp lệ")
	// }

	// 3. Kiểm tra mã phản hồi từ VNPAY
	if responseCode != "00" {
		return c.SendString("❌ Thanh toán bị từ chối hoặc thất bại")
	}

	ctx := context.Background()

	// 4. Tìm đơn thanh toán trong DB
	payment, err := h.PaymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("❌ Không tìm thấy đơn thanh toán")
	}

	if payment.Status == "success" {
		return c.SendString("✅ Đơn đã xử lý trước đó")
	}

	// 5. Cập nhật trạng thái đơn thanh toán
	if err := h.PaymentRepo.UpdateStatus(ctx, orderID, "success"); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Cập nhật trạng thái lỗi")
	}

	// 6. Tạo ticket cho user
	ticket := &models.Ticket{
		UserID:  payment.UserID,
		EventID: payment.EventID,
		// Price:   payment.Amount,
	}

	createdTicket, err := h.TicketRepo.CreateOne(ctx, payment.UserID, ticket)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Tạo vé thất bại")
	}

	// 7. ✅ Cập nhật TicketID vào đơn thanh toán
	err = h.PaymentRepo.UpdateTicketID(ctx, orderID, fmt.Sprintf("%d", createdTicket.ID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Gán TicketID vào Payment thất bại")
	}

	return c.SendString("✅ Thanh toán thành công, vé đã được tạo!")
}

// ✅ POST /payment/momo
func (h *PaymentHandler) CreateMomoCheckout(c *fiber.Ctx) error {
	type Body struct {
		EventID uint `json:"eventId"`
	}
	var body Body

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	userId := uint(c.Locals("userId").(float64))
	ctx := context.Background()

	event, err := h.EventRepo.GetOne(ctx, body.EventID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "event not found",
		})
	}

	orderID := utils.GenerateOrderID(userId, event.ID)

	payment := &models.Payment{
		OrderID: orderID,
		EventID: event.ID,
		UserID:  userId,
		Amount:  int(event.Price),
		Status:  "pending",
		Method:  "momo",
	}

	if err := h.PaymentRepo.Create(ctx, payment); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create payment",
		})
	}

	payURL, err := utils.CreateMomoPayment(orderID, payment.Amount) // bạn đã có hàm này
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate MoMo payment URL",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"url":    payURL,
	})
}

// ✅ GET /payment/momo-return
func (h *PaymentHandler) HandleMomoReturn(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	resultCode := c.Query("resultCode")

	fmt.Println("🔍 OrderID:", orderID)
	fmt.Println("🔍 ResultCode:", resultCode)

	if orderID == "" || resultCode == "" {
		return c.Status(http.StatusBadRequest).SendString("❌ Thiếu thông tin callback từ MoMo")
	}

	if resultCode != "0" {
		return c.SendString("❌ Thanh toán bị từ chối hoặc thất bại")
	}

	ctx := context.Background()

	payment, err := h.PaymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("❌ Không tìm thấy đơn thanh toán")
	}

	if payment.Status == "success" {
		return c.SendString("✅ Đơn đã xử lý trước đó")
	}

	if err := h.PaymentRepo.UpdateStatus(ctx, orderID, "success"); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Cập nhật trạng thái lỗi")
	}

	ticket := &models.Ticket{
		UserID:  payment.UserID,
		EventID: payment.EventID,
		// Price:   payment.Amount,
	}

	createdTicket, err := h.TicketRepo.CreateOne(ctx, payment.UserID, ticket)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Tạo vé thất bại")
	}

	err = h.PaymentRepo.UpdateTicketID(ctx, orderID, fmt.Sprintf("%d", createdTicket.ID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Gán TicketID vào Payment thất bại")
	}

	return c.SendString("✅ Thanh toán MoMo thành công, vé đã được tạo!")
}

// POST /payment/momo-ipn
func (h *PaymentHandler) HandleMomoIPN(c *fiber.Ctx) error {
	type MomoIPNBody struct {
		PartnerCode  string `json:"partnerCode"`
		OrderId      string `json:"orderId"`
		RequestId    string `json:"requestId"`
		Amount       string `json:"amount"`
		OrderInfo    string `json:"orderInfo"`
		OrderType    string `json:"orderType"`
		TransId      int64  `json:"transId"`
		ResultCode   int    `json:"resultCode"`
		Message      string `json:"message"`
		PayType      string `json:"payType"`
		ResponseTime int64  `json:"responseTime"`
		ExtraData    string `json:"extraData"`
		Signature    string `json:"signature"`
	}

	var body MomoIPNBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).SendString("❌ Dữ liệu không hợp lệ")
	}

	// Bước 1: Xác thực chữ ký (tương tự như tạo chữ ký khi gọi API)
	params := map[string]string{
		"amount":       body.Amount,
		"extraData":    body.ExtraData,
		"message":      body.Message,
		"orderId":      body.OrderId,
		"orderInfo":    body.OrderInfo,
		"orderType":    body.OrderType,
		"partnerCode":  body.PartnerCode,
		"payType":      body.PayType,
		"requestId":    body.RequestId,
		"responseTime": fmt.Sprintf("%d", body.ResponseTime),
		"resultCode":   fmt.Sprintf("%d", body.ResultCode),
		"transId":      fmt.Sprintf("%d", body.TransId),
	}

	// 🔐 Lấy secret từ env
	secretKey := os.Getenv("MOMO_SECRET_KEY")
	generatedSig := utils.GenerateMomoSignature(params, secretKey)
	
	if body.Signature != generatedSig {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Chữ ký không hợp lệ",
			"sign": generatedSig,
		})
	}

	// Bước 2: Kiểm tra kết quả giao dịch
	if body.ResultCode != 0 {
		return c.SendString("❌ Giao dịch thất bại hoặc bị hủy")
	}

	ctx := context.Background()

	// Bước 3: Kiểm tra đơn thanh toán
	payment, err := h.PaymentRepo.GetByOrderID(ctx, body.OrderId)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("❌ Không tìm thấy đơn thanh toán")
	}
	if payment.Status == "success" {
		return c.SendString("✅ Đơn đã được xử lý trước đó")
	}

	// Bước 4: Cập nhật đơn thanh toán
	err = h.PaymentRepo.UpdateStatus(ctx, body.OrderId, "success")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Cập nhật thanh toán lỗi")
	}

	// Bước 5: Tạo vé
	ticket := &models.Ticket{
		UserID:  payment.UserID,
		EventID: payment.EventID,
		// Price:   payment.Amount,
	}
	createdTicket, err := h.TicketRepo.CreateOne(ctx, payment.UserID, ticket)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Tạo vé lỗi")
	}

	// Bước 6: Gán TicketID vào đơn thanh toán
	err = h.PaymentRepo.UpdateTicketID(ctx, body.OrderId, fmt.Sprintf("%d", createdTicket.ID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("❌ Gán TicketID lỗi")
	}

	return c.SendString("✅ Giao dịch thành công, vé đã tạo!")
}

func NewPaymentHandler(router fiber.Router, pRepo models.PaymentRepository, eRepo models.EventRepository, tRepo models.TicketRepository) {
	handler := &PaymentHandler{
		PaymentRepo: pRepo,
		EventRepo:   eRepo,
		TicketRepo:  tRepo,
	}
	// momo
	router.Post("/momo", handler.CreateMomoCheckout)
	router.Get("/momo-return", handler.HandleMomoReturn)
	router.Post("/momo-ipn", handler.HandleMomoIPN)
	// vnpay
	router.Post("/vnpay", handler.CreateVnpayCheckout)
	router.Get("/vnpay-return", handler.HandleVnpayReturn)
}
