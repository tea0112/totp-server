package totp

import (
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func TestValidateWithUserDelay(t *testing.T) {
	tests := []struct {
		name        string
		period      uint
		fixedDelay  time.Duration
		shouldValid bool
	}{
		{
			name:        "user delay 5 seconds with 5min period",
			period:      300,
			fixedDelay:  5 * time.Second,
			shouldValid: true,
		},
		{
			name:        "user delay 30 seconds with 5min period",
			period:      300,
			fixedDelay:  30 * time.Second,
			shouldValid: true,
		},
		{
			name:        "user delay 5 seconds with 1min period",
			period:      60,
			fixedDelay:  5 * time.Second,
			shouldValid: true,
		},
		{
			name:        "user delay 30 seconds with 1min period",
			period:      60,
			fixedDelay:  30 * time.Second,
			shouldValid: true,
		},
		{
			name:        "user delay 60 seconds with 1min period (at boundary - new window)",
			period:      60,
			fixedDelay:  60 * time.Second,
			shouldValid: false,
		},
		{
			name:        "user delay 61 seconds with 1min period (new window)",
			period:      60,
			fixedDelay:  61 * time.Second,
			shouldValid: false,
		},
		{
			name:        "user delay 120 seconds with 1min period (2 windows later)",
			period:      60,
			fixedDelay:  120 * time.Second,
			shouldValid: false,
		},
		{
			name:        "user delay 180 seconds with 1min period (3 windows later)",
			period:      60,
			fixedDelay:  60 * 3 * time.Second,
			shouldValid: false,
		},
		{
			name:        "user delay 150 seconds with 5min period (same window)",
			period:      300,
			fixedDelay:  150 * time.Second,
			shouldValid: true,
		},
		{
			name:        "user delay 300 seconds with 5min period (at boundary)",
			period:      300,
			fixedDelay:  300 * time.Second,
			shouldValid: false,
		},
		{
			name:        "user delay 301 seconds with 5min period (new window)",
			period:      300,
			fixedDelay:  301 * time.Second,
			shouldValid: false,
		},
	}

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
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
			store.Set("test@example.com", secret)

			genTime := baseTime
			code, _ := service.GenerateCode(secret, genTime)

			validateTime := genTime.Add(tt.fixedDelay)
			valid, _ := service.ValidateAt("test@example.com", code, validateTime)

			if valid != tt.shouldValid {
				t.Errorf("delay=%v: expected valid=%v, got %v (period=%d)",
					tt.fixedDelay, tt.shouldValid, valid, tt.period)
			}
		})
	}
}

func TestValidateWithPeriod(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

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

			key, _ := totp.Generate(totp.GenerateOpts{
				Issuer:      "TOTPApp",
				AccountName: "test",
				Algorithm:   otp.AlgorithmSHA1,
				Digits:      otp.DigitsSix,
				Period:      tt.period,
				SecretSize:  20,
			})
			secret := key.Secret()
			store.Set("test@example.com", secret)

			genTime := baseTime
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
	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		setup       func(*Store) string
		code        string
		shouldValid bool
	}{
		{
			name: "wrong code",
			setup: func(s *Store) string {
				key, _ := totp.Generate(totp.GenerateOpts{
					Issuer:      "TOTPApp",
					AccountName: "test",
					Algorithm:   otp.AlgorithmSHA1,
					Digits:      otp.DigitsSix,
					Period:      300,
					SecretSize:  20,
				})
				secret := key.Secret()
				s.Set("test@example.com", secret)
				return "000000"
			},
			code:        "000000",
			shouldValid: false,
		},
		{
			name: "non-existent account",
			setup: func(s *Store) string {
				key, _ := totp.Generate(totp.GenerateOpts{
					Issuer:      "TOTPApp",
					AccountName: "other",
					Algorithm:   otp.AlgorithmSHA1,
					Digits:      otp.DigitsSix,
					Period:      300,
					SecretSize:  20,
				})
				secret := key.Secret()
				s.Set("other@example.com", secret)
				code, _ := totp.GenerateCodeCustom(secret, baseTime, totp.ValidateOpts{
					Period:    300,
					Digits:    otp.DigitsSix,
					Algorithm: otp.AlgorithmSHA1,
				})
				return code
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
	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

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

			code, _ := service.GenerateCode(secret, baseTime)

			if len(code) != 6 {
				t.Errorf("expected 6 digits, got %d", len(code))
			}
		})
	}
}
