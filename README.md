# Ticket Booking App - Backend

## Hướng dẫn cài đặt và chạy dự án

### 1. Cài đặt Go SDK

- Truy cập [https://go.dev/dl/](https://go.dev/dl/) để tải và cài đặt Go phiên bản >= 1.22.2.
- Kiểm tra cài đặt bằng lệnh: `go version`

### 2. Clone repository về máy
```bash
git clone https://github.com/minq3010/Ticket-Booking-App-Backend.git
cd Ticket-Booking-App-Backend/Backend
```
### 3. Tạo nhánh mới
```bash
git checkout -b <ten-nhanh-cua-ban>
```
### 4. Cài đặt các package cần thiết
```bash
go mod tidy
```
### 5. Đã có sẵn file môi trường

- Cần tạo lại file `.env` (Liên hệ)

### 6. Chạy dự án bằng Makefile (Docker)
```bash
make run 
```
### 7. Dừng dự án
```bash
make stop
Ctrl C
```

### 8. Pull code mới nhất từ remote
```bash
git pull origin main
```
### 9. Push code lên remote
```bash
git add .
git commit -m "Mô tả thay đổi"
git push origin <ten-nhanh-cua-ban>
```
**Lưu ý:**  
- Đảm bảo đã cài đặt Go, Docker và Make trên máy.
- Kiểm tra các yêu cầu khác trong file `go.mod` hoặc tài liệu dự án.
