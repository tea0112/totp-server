package main

import (
	"log"
	"net/http"

	"totp-app/config"
	"totp-app/totp"
)

func main() {
	cfg := config.Load()

	store := totp.NewStore()
	service := totp.NewService(store, uint(cfg.TOTPPeriod.Seconds()))
	handler := totp.NewHandler(service)

	http.HandleFunc("/totp/generate", handler.GenerateTOTP)
	http.HandleFunc("/totp/verify", handler.VerifyTOTP)
	http.HandleFunc("/health", healthCheck)

	log.Printf("Server starting on :%s", cfg.ServerPort)
	log.Printf("TOTP period: %d seconds", int(cfg.TOTPPeriod.Seconds()))
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
