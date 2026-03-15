# TOTP Server

TOTP (Time-based One-Time Password) server với Go và thư viện [pquerna/otp](https://github.com/pquerna/otp).

## Mục đích

Ứng dụng mô phỏng việc tạo và xác thực mã TOTP với các đặc điểm:

- **Period tùy chỉnh**: 5 phút (300 giây) - mã sống trong 5 phút sau khi tạo
- **Không có window lệch**: Code chỉ valid trong cùng window với thời điểm tạo
- **In-memory storage**: Lưu secret trong memory
- **Config qua environment**: Dễ dàng cấu hình period và port

## Cấu trúc project

```
.
├── .env                    # Environment variables
├── Makefile               # Commands tiện lợi
├── main.go               # Entry point
├── config/
│   └── config.go         # Load configuration
├── totp/
│   ├── types.go         # Request/Response models
│   ├── store.go         # In-memory storage
│   ├── service.go       # Business logic
│   ├── handler.go       # HTTP handlers
│   └── validate_test.go # Unit tests (table-driven)
└── docs/
    └── TOTP_EXPLAINED.md # Giải thích chi tiết về TOTP
```

## API Endpoints

| Method | Endpoint       | Mô tả                           |
|--------|----------------|--------------------------------|
| POST   | `/totp/generate` | Tạo TOTP code mới, log ra console |
| POST   | `/totp/verify`   | Verify TOTP code                |
| GET    | `/health`        | Health check                   |

## Cấu hình

File `.env`:

```bash
TOTP_PERIOD=300    # Thời gian sống của mã (giây). VD: 300 = 5 phút
SERVER_PORT=8080   # Port chạy server
```

## Sử dụng

### Chạy server

```bash
make run
# Hoặc
go run main.go
```

### Generate TOTP

```bash
curl -X POST http://localhost:8080/totp/generate \
  -H "Content-Type: application/json" \
  -d '{"account_name": "user@example.com"}'

# Response: {"message": "TOTP generated successfully. Check server logs for the code."}
# Server log: [TOTP] Code at 10:00:00: 123456 (valid for 300 seconds)
```

### Verify TOTP

```bash
curl -X POST http://localhost:8080/totp/verify \
  -H "Content-Type: application/json" \
  -d '{"account_name": "user@example.com", "code": "123456"}'

# Response: {"valid": true}
```

### Các commands khác

```bash
make test       # Chạy tất cả tests
make build      # Build binary
make clean      # Xóa binary
make health     # Health check
```

## Tests

Chạy tests:

```bash
make test
# Hoặc
go test ./...
```

### Các test cases

- `TestValidateWithUserDelay` - Test với độ trễ người dùng (5s, 30s, 60s, ...)
- `TestValidateWithPeriod` - Test validate với các period khác nhau
- `TestValidateFailureCases` - Test các trường hợp fail (sai code, account không tồn tại)
- `TestGenerateCodeLength` - Test độ dài code (6 digits)

Xem chi tiết trong `docs/TOTP_EXPLAINED.md`.

## Tài liệu tham khảo

- [RFC 6238 - TOTP](https://tools.ietf.org/html/rfc6238)
- [pquerna/otp](https://github.com/pquerna/otp)
- [docs/TOTP_EXPLAINED.md](./docs/TOTP_EXPLAINED.md) - Giải thích chi tiết về cách TOTP hoạt động
