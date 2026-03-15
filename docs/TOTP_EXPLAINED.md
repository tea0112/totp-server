# TOTP (Time-based One-Time Password) Explained

## 1. Cách TOTP hoạt động

### Công thức TOTP

```
TOTP = HMAC-SHA1(secret, floor(unix_timestamp / period))
```

**Trong đó:**
- `secret`: Khóa bí mật được tạo khi đăng ký TOTP
- `unix_timestamp`: Thời gian hiện tại tính bằng giây
- `period`: Khoảng thời gian mỗi cửa sổ (ví dụ: 60 giây, 300 giây)
- `floor(x)`: Lấy phần nguyên

### Time Window (Cửa sổ thời gian)

TOTP chia thời gian thành các "cửa sổ" (windows) có kích thước bằng `period`:

```
Window = floor(unix_timestamp / period)
```

**Ví dụ với period = 60 giây:**

| Thời gian | Unix Timestamp | Window | Mã TOTP |
|-----------|----------------|--------|---------|
| 15:45:00 | X | floor(X/60) = 100 | ABC123 |
| 15:45:30 | X+30 | floor((X+30)/60) = 100 | ABC123 |
| 15:45:59 | X+59 | floor((X+59)/60) = 100 | ABC123 |
| 15:46:00 | X+60 | floor((X+60)/60) = 101 | DEF456 |
| 15:46:30 | X+90 | floor((X+90)/60) = 101 | DEF456 |

**Quan sát:** Tất cả thời điểm trong cùng 1 phút (window 100) tạo ra cùng một mã TOTP!

---

## 2. Giải thích test case

### Code trong test (sử dụng time cố định):

```go
baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

// Server tạo TOTP tại thời điểm baseTime
service.Generate("test@example.com")
secret, _ := store.Get("test@example.com")

// Tạo lại code với cùng thời điểm baseTime
code, _ := service.GenerateCode(secret, baseTime)
```

### Tại sao 2 lần gọi tạo ra CÙNG một mã?

- `service.Generate()` bên trong gọi `GenerateCodeCustom(secret, time.Now(), ...)`
- Nhưng chúng ta sử dụng `baseTime` cố định để tạo lại code
- **Cùng thời điểm** → **Cùng window** → **Cùng mã TOTP**

**Minh họa với baseTime = 10:00:00:**

```
Tạo lúc: 10:00:00 → window = floor(10:00:00 / 60) = 1000
Code: ABC123

Tạo lại với baseTime: 10:00:00 → window = floor(10:00:00 / 60) = 1000
Code: ABC123 (TRÙNG!)
```

---

## 3. Tại sao thiết kế này?

### Lý do TOTP sinh cùng mã trong 1 window:

1. **Đồng bộ thời gian:** Client và server có thể không đồng bộ hoàn toàn (± vài giây). Miễn là cùng window, mã vẫn hợp lệ.

2. **Thuận tiện cho user:** User có ~30-60 giây để nhập mã (với period=60) thay vì chỉ 1 giây chính xác.

3. **Không cần đồng bộ exact:** Server chỉ cần kiểm tra mã với window tương ứng.

---

## 4. Các test cases trong file

Sử dụng `time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)` để đảm bảo deterministic.

### Test: `valid at exact creation time`
```go
baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
code := GenerateCode(secret, baseTime)  // window = 1000, mã = ABC123
valid := ValidateAt(code, baseTime)     // verify tại 10:00:00 → window = 1000 → ✓ VALID
```

### Test: `invalid at 6min after creation with period=300`
```go
baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
code := GenerateCode(secret, baseTime)    // window = floor(10:06:00 / 300) = 20
validateTime := baseTime.Add(6 * time.Minute)
valid := ValidateAt(code, validateTime) // verify tại 10:06:00 → window = 21 → ✗ INVALID
```

### Test: `verify at 30sec with period 60 - same window`
```go
baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
code := GenerateCode(secret, baseTime)    // window = floor(10:00:30 / 60) = 1000
validateTime := baseTime.Add(30 * time.Second)
valid := ValidateAt(code, validateTime)  // verify tại 10:00:30 → window = 1000 → ✓ VALID (cùng window)
```

---

## 5. Tóm tắt

| Khái niệm | Giải thích |
|-----------|------------|
| **Period** | Thời gian mỗi cửa sổ (60s, 300s,...) |
| **Window** | `floor(timestamp / period)` - đánh số mỗi cửa sổ |
| **TOTP Code** | HMAC-SHA1(secret, window) - chỉ phụ thuộc vào window |
| **Cùng window** | → Cùng mã TOTP |
| **Khác window** | → Mã TOTP khác nhau |

**Điều quan trọng:** Mã TOTP không phụ thuộc vào "khoảng cách thời gian" mà phụ thuộc vào "cửa sổ thời gian" (window).

---

## 6. User Delay (Độ trễ người dùng)

Trong thực tế, người dùng cần thời gian để mở email/SMS và nhập mã TOTP. Điều này được mô phỏng trong test bằng "user delay".

### Flow thực tế:

```
Server tạo code tại T      → Gửi email/SMS cho user
User đợi delay (5-30s)     → Mở email, đọc mã
User submit code tại T+delay → Server validate tại T+delay
```

### Code test (sử dụng time cố định):

```go
baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

// Time T: Server tạo và gửi code cho user
code, _ := service.GenerateCode(secret, baseTime)

// Time T + delay: User nhập code sau khi đợi delay
validateTime := baseTime.Add(tt.fixedDelay)
valid, _ := service.ValidateAt("test@example.com", code, validateTime)
```

### Các test cases:

```go
// Period = 60 giây (1 phút)
- delay 5s   → cùng window → VALID
- delay 60s  → window mới → INVALID
- delay 61s  → window mới → INVALID

// Period = 300 giây (5 phút)
- delay 150s  → cùng window → VALID
- delay 300s  → window mới → INVALID
- delay 301s  → window mới → INVALID
```

### Quan trọng:

- **Period = 60s**: Mã chỉ có hiệu lực trong vòng 60 giây (1 window). Nếu user mất 30 giây để nhập mã (và không cross boundary), vẫn hợp lệ.
- **Period = 300s**: Mã có hiệu lực trong 5 phút. User có nhiều thời gian hơn.

### Lưu ý về window boundary:

Nếu thời điểm tạo mã gần cuối window (ví dụ: giây thứ 55 với period=60), thì chỉ cần 5 giây sau đã sang window mới!

```
Tạo lúc: 15:45:55 → window = floor(15:45:55 / 60) = 945
Verify lúc: 15:46:00 → window = floor(15:46:00 / 60) = 946 → INVALID!
```

Đây là lý do tại sao test case "delay 30s với period=60" có thể fail tùy thuộc vào thời điểm chính xác của `time.Now()`.
