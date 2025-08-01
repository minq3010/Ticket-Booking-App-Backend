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

//  POST /payment/momo
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
		"status":  "success",
		"url":     payURL,
		"orderID": orderID,
		"message": "MoMo payment URL created successfully",
	})
}

//  GET /payment/momo-return
func (h *PaymentHandler) HandleMomoReturn(c *fiber.Ctx) error {
	//  Debug tất cả query parameters
	fmt.Printf("🔍 All query params: %+v\n", c.Queries())

	// Thử nhiều cách lấy orderID
	orderID := c.Query("order_id")
	if orderID == "" {
		orderID = c.Query("orderId")
	}
	if orderID == "" {
		orderID = c.Query("orderInfo")
	}

	resultCode := c.Query("resultCode")
	if resultCode == "" {
		resultCode = c.Query("result_code")
	}

	fmt.Printf("🔍 MoMo Callback - OrderID: '%s', ResultCode: '%s'\n", orderID, resultCode)

	//  Nếu thiếu parameters
	if orderID == "" || resultCode == "" {
		fmt.Printf("Missing parameters - OrderID: '%s', ResultCode: '%s'\n", orderID, resultCode)
		return utils.RenderPaymentError(c, utils.PaymentPageData{
			Title:      "Lỗi thanh toán",
			Heading:    "Lỗi thanh toán!",
			Message:    "Không thể xử lý thông tin thanh toán.",
			SubMessage: "Vui lòng thử lại hoặc liên hệ hỗ trợ.",
		})
	}

	//  Kiểm tra result code trước
	if resultCode != "0" {
		fmt.Printf("Payment failed with code: %s\n", resultCode)
		return utils.RenderPaymentError(c, utils.PaymentPageData{
			Title:      "Thanh toán thất bại",
			Heading:    "Thanh toán thất bại!",
			Message:    "Giao dịch không thành công.",
			SubMessage: "Vui lòng thử lại hoặc chọn phương thức thanh toán khác.",
			ErrorCode:  resultCode,
		})
	}

	ctx := context.Background()
	payment, err := h.PaymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		fmt.Printf("Payment not found: %v\n", err)
		return utils.RenderPaymentError(c, utils.PaymentPageData{
			Title:      "Không tìm thấy đơn hàng",
			Heading:    "Không tìm thấy đơn hàng!",
			Message:    "Đơn hàng không tồn tại trong hệ thống.",
			SubMessage: "Vui lòng liên hệ hỗ trợ khách hàng.",
			OrderID:    orderID,
		})
	}

	//  Tránh xử lý duplicate
	if payment.Status == "success" {
		fmt.Println(" Payment already processed, showing success page")
		return utils.RenderPaymentSuccess(c, utils.PaymentPageData{
			Heading:    "Thanh toán thành công!",
			Message:    "Vé đã được tạo thành công.",
			SubMessage: "Vui lòng quay về app để xem vé của bạn.",
			OrderID:    orderID,
		})
	}

	// Update payment status
	if err := h.PaymentRepo.UpdateStatus(ctx, orderID, "success"); err != nil {
		fmt.Printf("Failed to update payment status: %v\n", err)
		return utils.RenderPaymentError(c, utils.PaymentPageData{
			Title:      "Lỗi hệ thống",
			Heading:    "Lỗi hệ thống!",
			Message:    "Không thể cập nhật trạng thái thanh toán.",
			SubMessage: "Vui lòng liên hệ hỗ trợ khách hàng.",
		})
	}

	//  Create ticket
	ticket := &models.Ticket{
		UserID:  payment.UserID,
		EventID: payment.EventID,
	}

	createdTicket, err := h.TicketRepo.CreateOne(ctx, payment.UserID, ticket)
	if err != nil {
		fmt.Printf("❌ Failed to create ticket: %v\n", err)
		return utils.RenderPaymentError(c, utils.PaymentPageData{
			Title:      "Lỗi tạo vé",
			Heading:    "Lỗi tạo vé!",
			Message:    "Thanh toán thành công nhưng không thể tạo vé.",
			SubMessage: "Vui lòng liên hệ hỗ trợ khách hàng để được hỗ trợ.",
			OrderID:    orderID,
		})
	}

	//  Link ticket to payment
	if err := h.PaymentRepo.UpdateTicketID(ctx, orderID, fmt.Sprintf("%d", createdTicket.ID)); err != nil {
		fmt.Printf("Failed to link ticket: %v\n", err)
		// Ticket đã tạo thành công, chỉ log warning
		fmt.Println("️Ticket created but not linked to payment")
	}

	fmt.Printf("Payment successful - OrderID: %s, TicketID: %d\n", orderID, createdTicket.ID)

	//  Success page cuối cùng
	return utils.RenderPaymentSuccess(c, utils.PaymentPageData{
		Heading:    "Thanh toán thành công!",
		Message:    "Chúc mừng! Vé của bạn đã được tạo thành công.",
		SubMessage: "Vui lòng quay về app để xem và sử dụng vé của bạn.",
		OrderID:    orderID,
		TicketID:   fmt.Sprintf("%d", createdTicket.ID),
	})
}

// POST /payment/momo-ipn
func (h *PaymentHandler) HandleMomoIPN(c *fiber.Ctx) error {
	type MomoIPNBody struct {
		PartnerCode  string `json:"partnerCode"`
		OrderId      string `json:"orderId"`
		RequestId    string `json:"requestId"`
		Amount       int64  `json:"amount"`
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
		fmt.Println("IPN Body parse error:", err)
		return c.Status(http.StatusBadRequest).SendString("Dữ liệu không hợp lệ")
	}

	fmt.Printf("🔍 IPN Received: OrderID=%s, ResultCode=%d, Signature=%s\n",
		body.OrderId, body.ResultCode, body.Signature)

	// Bước 1: Xác thực chữ ký với format IPN
	params := map[string]string{
		"accessKey":    os.Getenv("MOMO_ACCESS_KEY"),
		"amount":       fmt.Sprintf("%d", body.Amount),
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

	secretKey := os.Getenv("MOMO_SECRET_KEY")
	generatedSig := utils.GenerateMomoIPNSignature(params, secretKey) 

	fmt.Printf(" Generated signature: %s\n", generatedSig)
	fmt.Printf(" Received signature:  %s\n", body.Signature)

	if body.Signature != generatedSig {
		fmt.Println(" Signature verification failed!")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":     "Chữ ký không hợp lệ",
			"received":  body.Signature,
			"generated": generatedSig,
		})
	}

	fmt.Println("Signature verified successfully!")

	// Bước 2: Kiểm tra kết quả giao dịch
	if body.ResultCode != 0 {
		fmt.Printf("Transaction failed with code: %d\n", body.ResultCode)
		return c.SendString("Giao dịch thất bại hoặc bị hủy")
	}

	ctx := context.Background()

	//  Bước 3: Kiểm tra đơn thanh toán
	payment, err := h.PaymentRepo.GetByOrderID(ctx, body.OrderId)
	if err != nil {
		fmt.Printf(" Payment not found: %v\n", err)
		return c.Status(http.StatusNotFound).SendString(" Không tìm thấy đơn thanh toán")
	}

	if payment.Status == "success" {
		fmt.Println("Payment already processed")
		return c.SendString("Đơn đã được xử lý trước đó")
	}

	//  Bước 4: Cập nhật đơn thanh toán
	err = h.PaymentRepo.UpdateStatus(ctx, body.OrderId, "success")
	if err != nil {
		fmt.Printf(" Failed to update payment: %v\n", err)
		return c.Status(http.StatusInternalServerError).SendString("Cập nhật thanh toán lỗi")
	}

	// Bước 5: Tạo vé
	ticket := &models.Ticket{
		UserID:  payment.UserID,
		EventID: payment.EventID,
	}
	createdTicket, err := h.TicketRepo.CreateOne(ctx, payment.UserID, ticket)
	if err != nil {
		fmt.Printf("Failed to create ticket: %v\n", err)
		return c.Status(http.StatusInternalServerError).SendString("Tạo vé lỗi")
	}

	// Bước 6: Gán TicketID vào đơn thanh toán
	err = h.PaymentRepo.UpdateTicketID(ctx, body.OrderId, fmt.Sprintf("%d", createdTicket.ID))
	if err != nil {
		fmt.Printf(" Failed to link ticket: %v\n", err)
		return c.Status(http.StatusInternalServerError).SendString(" Gán TicketID lỗi")
	}

	fmt.Printf("IPN processed successfully - OrderID: %s, TicketID: %d\n", body.OrderId, createdTicket.ID)
	return c.SendString("Giao dịch thành công, vé đã tạo!")
}

// GET /payment/status/:orderID
func (h *PaymentHandler) GetPaymentStatus(c *fiber.Ctx) error {
	orderID := c.Params("orderID")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OrderID is required",
		})
	}

	ctx := context.Background()
	payment, err := h.PaymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Payment not found",
		})
	}

	return c.JSON(fiber.Map{
		"orderID": payment.OrderID,
		"status":  payment.Status,
		"amount":  payment.Amount,
	})
}

func NewPaymentHandler(router fiber.Router, pRepo models.PaymentRepository, eRepo models.EventRepository, tRepo models.TicketRepository) {
	handler := &PaymentHandler{
		PaymentRepo: pRepo,
		EventRepo:   eRepo,
		TicketRepo:  tRepo,
	}
	// momo
	router.Post("/momo", handler.CreateMomoCheckout)
	router.Get("/status/:orderID", handler.GetPaymentStatus)
}

func NewPaymentCallbackHandler(router fiber.Router, pRepo models.PaymentRepository, eRepo models.EventRepository, tRepo models.TicketRepository) {
	handler := &PaymentHandler{
		PaymentRepo: pRepo,
		EventRepo:   eRepo,
		TicketRepo:  tRepo,
	}
	// GET/POST routes KHÔNG cần authentication
	router.Get("/momo-return", handler.HandleMomoReturn)
	router.Post("/momo-ipn", handler.HandleMomoIPN)
}
