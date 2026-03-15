package totp

type GenerateRequest struct {
	AccountName string `json:"account_name"`
}

type GenerateResponse struct {
	Message string `json:"message"`
}

type VerifyRequest struct {
	AccountName string `json:"account_name"`
	Code        string `json:"code"`
}

type VerifyResponse struct {
	Valid bool `json:"valid"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
