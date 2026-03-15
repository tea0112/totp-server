package totp

import (
	"log"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type Service struct {
	store  *Store
	period uint
}

func NewService(store *Store, period uint) *Service {
	if period == 0 {
		period = 300
	}
	return &Service{
		store:  store,
		period: period,
	}
}

func (s *Service) Generate(accountName string) error {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "TOTPApp",
		AccountName: accountName,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      s.period,
		SecretSize:  20,
	})
	if err != nil {
		return err
	}

	secret := key.Secret()
	isNew := s.store.Set(accountName, secret)

	now := time.Now()
	code, _ := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    s.period,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if isNew {
		log.Printf("[TOTP] Created new TOTP for account: %s", accountName)
	} else {
		log.Printf("[TOTP] Updated TOTP for account: %s", accountName)
	}
	log.Printf("[TOTP] Code at %s: %s (valid for %d seconds)", now.Format("15:04:05"), code, s.period)

	return nil
}

func (s *Service) GenerateCode(secret string, t time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, t, totp.ValidateOpts{
		Period:    s.period,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
}

func (s *Service) Validate(accountName, code string) (bool, error) {
	secret, exists := s.store.Get(accountName)
	if !exists {
		return false, nil
	}

	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    s.period,
			Skew:      0,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	return valid, err
}

func (s *Service) ValidateAt(accountName, code string, t time.Time) (bool, error) {
	secret, exists := s.store.Get(accountName)
	if !exists {
		return false, nil
	}

	valid, err := totp.ValidateCustom(
		code,
		secret,
		t,
		totp.ValidateOpts{
			Period:    s.period,
			Skew:      0,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	return valid, err
}

func (s *Service) GetSecret(accountName string) (string, bool) {
	return s.store.Get(accountName)
}
