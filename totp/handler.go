package totp

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GenerateTOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body")
		return
	}

	if req.AccountName == "" {
		h.writeError(w, "account_name is required")
		return
	}

	err := h.service.Generate(req.AccountName)
	if err != nil {
		h.writeError(w, "Failed to generate TOTP: "+err.Error())
		return
	}

	resp := GenerateResponse{Message: "TOTP generated successfully. Check server logs for the code."}
	h.writeJSON(w, resp)
}

func (h *Handler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body")
		return
	}

	if req.AccountName == "" || req.Code == "" {
		h.writeError(w, "account_name and code are required")
		return
	}

	valid, err := h.service.Validate(req.AccountName, req.Code)
	if err != nil {
		h.writeError(w, "Failed to verify TOTP: "+err.Error())
		return
	}

	resp := VerifyResponse{Valid: valid}
	h.writeJSON(w, resp)
}

func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
