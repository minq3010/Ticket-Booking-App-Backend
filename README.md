# Ticket Booking App - Backend

## Hướng dẫn cài đặt và chạy dự án bằng Docker

### 1. Clone repository về máy
```bash
git clone https://github.com/minq3010/Ticket-Booking-App-Backend.git
cd Ticket-Booking-App-Backend
```

### 2. Tạo nhánh mới
```bash
git checkout -b <ten-nhanh-cua-ban>
```

### 3. Chuẩn bị file môi trường
- Cần tạo file `.env` (Liên hệ để nhận file mẫu nếu chưa có)

### 4. Chạy dự án bằng Docker Compose
```bash
docker compose up --build
```

### 5. Dừng dự án
```bash
docker compose down
```

### 6. Pull code mới nhất từ remote
```bash
git pull origin main
```

### 7. Push code lên remote
```bash
git add .
git commit -m "Mô tả thay đổi"
git push origin <ten-nhanh-cua-ban>
```

**Lưu ý:**  
- Đảm bảo đã cài đặt Docker và Docker Compose trên máy.
- Không cần cài đặt Go SDK trên máy local.
- Kiểm tra các yêu cầu khác trong file `go.mod` hoặc tài liệu dự án nếu cần.
