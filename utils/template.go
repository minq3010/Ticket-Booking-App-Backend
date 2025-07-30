package utils

import (
	"bytes"
	"html/template"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type PaymentPageData struct {
	Title      string
	Heading    string
	Message    string
	SubMessage string
	OrderID    string
	TicketID   string
	ErrorCode  string
}

func RenderPaymentError(c *fiber.Ctx, data PaymentPageData) error {
	tmplPath := filepath.Join("templates", "payment", "error.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		// Fallback to simple HTML if template fails
		fallbackHTML := `
		<html>
		<head><title>Lỗi thanh toán</title><meta charset="UTF-8"></head>
		<body style="text-align:center;padding:50px;font-family:Arial;">
			<h1>❌ ` + data.Heading + `</h1>
			<p>` + data.Message + `</p>
			<p>` + data.SubMessage + `</p>
			<button onclick="window.close()">Đóng trang này</button>
			<script>setTimeout(() => window.close(), 5000);</script>
		</body>
		</html>`
		return c.Type("html").SendString(fallbackHTML)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Template execution error")
	}

	return c.Type("html").SendString(buf.String())
}

func RenderPaymentSuccess(c *fiber.Ctx, data PaymentPageData) error {
	tmplPath := filepath.Join("templates", "payment", "success.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		// Fallback to simple HTML if template fails
		fallbackHTML := `
		<html>
		<head><title>Thanh toán thành công</title><meta charset="UTF-8"></head>
		<body style="text-align:center;padding:50px;font-family:Arial;">
			<h1>✅ ` + data.Heading + `</h1>
			<p>` + data.Message + `</p>
			<p>Mã đơn hàng: ` + data.OrderID + `</p>
			<button onclick="window.close()">Đóng trang này</button>
			<script>setTimeout(() => window.close(), 8000);</script>
		</body>
		</html>`
		return c.Type("html").SendString(fallbackHTML)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Template execution error")
	}

	return c.Type("html").SendString(buf.String())
}
