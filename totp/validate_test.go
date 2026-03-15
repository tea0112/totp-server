package totp

import (
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func TestValidateWithPeriod(t *testing.T) {
	tests := []struct {
		name        string
		period      uint
		timeOffset  time.Duration
		shouldValid bool
	}{
		{
			name:        "valid at exact creation time",
			period:      300,
			timeOffset:  0,
			shouldValid: true,
		},
		{
			name:        "invalid at 6min after creation (outside period)",
			period:      300,
			timeOffset:  6 * time.Minute,
			shouldValid: false,
		},
		{
			name:        "valid at exact creation time with 1min period",
			period:      60,
			timeOffset:  0,
			shouldValid: true,
		},
		{
			name:        "invalid at 2min after creation with 1min period",
			period:      60,
			timeOffset:  2 * time.Minute,
			shouldValid: false,
		},
		{
			name:        "invalid at 61sec after creation with 1min period",
			period:      60,
			timeOffset:  61 * time.Second,
			shouldValid: false,
		},
		{
			name:        "verify at 30sec with period 60 - same window",
			period:      60,
			timeOffset:  30 * time.Second,
			shouldValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			service := NewService(store, tt.period)

			genTime := time.Now()
			service.Generate("test@example.com")
			secret, _ := store.Get("test@example.com")

			code, _ := service.GenerateCode(secret, genTime)

			validateTime := genTime.Add(tt.timeOffset)
			valid, _ := service.ValidateAt("test@example.com", code, validateTime)

			if valid != tt.shouldValid {
				t.Errorf("expected valid=%v, got %v (period=%d, offset=%v)",
					tt.shouldValid, valid, tt.period, tt.timeOffset)
			}
		})
	}
}

func TestValidateFailureCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Store) string
		code        string
		shouldValid bool
	}{
		{
			name: "wrong code",
			setup: func(s *Store) string {
				service := NewService(s, 300)
				service.Generate("test@example.com")
				return "000000"
			},
			code:        "000000",
			shouldValid: false,
		},
		{
			name: "non-existent account",
			setup: func(s *Store) string {
				service := NewService(s, 300)
				service.Generate("other@example.com")
				return "123456"
			},
			code:        "123456",
			shouldValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			service := NewService(store, 300)
			tt.code = tt.setup(store)

			valid, _ := service.Validate("test@example.com", tt.code)
			if valid != tt.shouldValid {
				t.Errorf("expected valid=%v, got %v", tt.shouldValid, valid)
			}
		})
	}
}

func TestGenerateCodeLength(t *testing.T) {
	tests := []struct {
		name   string
		period uint
	}{
		{"6 digits with 5min period", 300},
		{"6 digits with 1min period", 60},
		{"6 digits with 30sec period", 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			service := NewService(store, tt.period)

			key, _ := totp.Generate(totp.GenerateOpts{
				Issuer:      "TOTPApp",
				AccountName: "test",
				Algorithm:   otp.AlgorithmSHA1,
				Digits:      otp.DigitsSix,
				Period:      tt.period,
				SecretSize:  20,
			})
			secret := key.Secret()

			code, _ := service.GenerateCode(secret, time.Now())

			if len(code) != 6 {
				t.Errorf("expected 6 digits, got %d", len(code))
			}
		})
	}
}
