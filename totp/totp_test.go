package totp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGenerateTOTP(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)

	err := service.Generate("test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate TOTP: %v", err)
	}

	stored, exists := store.Get("test@example.com")
	if !exists {
		t.Error("Secret should exist in store")
	}
	if stored == "" {
		t.Error("Secret should not be empty")
	}
}

func TestGenerateCode(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)

	service.Generate("test@example.com")
	secret, _ := store.Get("test@example.com")
	now := time.Now()

	code, err := service.GenerateCode(secret, now)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Code length should be 6, got %d", len(code))
	}

	valid, err := service.ValidateAt("test@example.com", code, now)
	if err != nil {
		t.Fatalf("Failed to validate: %v", err)
	}
	if !valid {
		t.Error("Code should be valid at generation time")
	}
}

func TestValidateSuccess(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)

	service.Generate("test@example.com")
	secret, _ := store.Get("test@example.com")
	now := time.Now()

	code, _ := service.GenerateCode(secret, now)
	valid, err := service.Validate("test@example.com", code)
	if err != nil {
		t.Fatalf("Failed to validate: %v", err)
	}
	if !valid {
		t.Error("Code should be valid")
	}
}

func TestValidateFailure(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)

	service.Generate("test@example.com")

	valid, err := service.Validate("test@example.com", "000000")
	if err != nil {
		t.Fatalf("Failed to validate: %v", err)
	}
	if valid {
		t.Error("Invalid code should not be valid")
	}
}

func TestValidateExpired(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)

	service.Generate("test@example.com")
	secret, _ := store.Get("test@example.com")
	now := time.Now()

	code, err := service.GenerateCode(secret, now)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	valid, err := service.Validate("test@example.com", code)
	if err != nil {
		t.Fatalf("Failed to validate: %v", err)
	}
	if !valid {
		t.Error("Code should be valid at generation time")
	}

	laterTime := now.Add(-6 * time.Minute)
	code2, _ := service.GenerateCode(secret, laterTime)

	valid2, _ := service.ValidateAt("test@example.com", code2, now)
	if valid2 {
		t.Error("Code should be invalid after 5 minutes")
	}
}

func TestValidateNonExistentAccount(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)

	valid, err := service.Validate("nonexistent@example.com", "000000")
	if err != nil {
		t.Fatalf("Failed to validate: %v", err)
	}
	if valid {
		t.Error("Code should not be valid for non-existent account")
	}
}

func TestStoreSetGet(t *testing.T) {
	store := NewStore()

	store.Set("user1", "secret1")
	store.Set("user2", "secret2")

	secret, exists := store.Get("user1")
	if !exists {
		t.Error("Secret should exist")
	}
	if secret != "secret1" {
		t.Errorf("Expected secret1, got %s", secret)
	}

	secret, exists = store.Get("user2")
	if !exists {
		t.Error("Secret should exist")
	}
	if secret != "secret2" {
		t.Errorf("Expected secret2, got %s", secret)
	}
}

func TestStoreSetNew(t *testing.T) {
	store := NewStore()

	isNew := store.Set("user1", "secret1")
	if !isNew {
		t.Error("Should be new when first set")
	}

	isNew = store.Set("user1", "secret2")
	if isNew {
		t.Error("Should not be new when updating")
	}
}

func TestStoreDelete(t *testing.T) {
	store := NewStore()

	store.Set("user1", "secret1")
	deleted := store.Delete("user1")
	if !deleted {
		t.Error("Should delete existing key")
	}

	_, exists := store.Get("user1")
	if exists {
		t.Error("Secret should not exist after delete")
	}

	deleted = store.Delete("nonexistent")
	if deleted {
		t.Error("Should not delete non-existent key")
	}
}

func TestHandlerGenerate(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)
	handler := NewHandler(service)

	body := `{"account_name": "test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/totp/generate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.GenerateTOTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp GenerateResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestHandlerGenerateInvalid(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)
	handler := NewHandler(service)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/totp/generate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.GenerateTOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandlerVerify(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)
	handler := NewHandler(service)

	service.Generate("test@example.com")
	secret, _ := store.Get("test@example.com")
	code, _ := service.GenerateCode(secret, time.Now())

	body := `{"account_name": "test@example.com", "code": "` + code + `"}`
	req := httptest.NewRequest(http.MethodPost, "/totp/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.VerifyTOTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp VerifyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Valid {
		t.Error("Code should be valid")
	}
}

func TestHandlerVerifyFailure(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)
	handler := NewHandler(service)

	service.Generate("test@example.com")

	body := `{"account_name": "test@example.com", "code": "000000"}`
	req := httptest.NewRequest(http.MethodPost, "/totp/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.VerifyTOTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp VerifyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Valid {
		t.Error("Code should not be valid")
	}
}

func TestHandlerVerifyInvalid(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)
	handler := NewHandler(service)

	body := `{"account_name": "test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/totp/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.VerifyTOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	store := NewStore()
	service := NewService(store, 300)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/totp/generate", nil)
	w := httptest.NewRecorder()

	handler.GenerateTOTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
