package handler

import (
	"encoding/json"
	"net/http"

	"imgbed/auth"

	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

type LoginHandler struct {
	passwordHash string
}

func NewLoginHandler(passwordHash string) *LoginHandler {
	return &LoginHandler{passwordHash: passwordHash}
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, `{"error":"password required"}`, http.StatusBadRequest)
		return
	}

	if h.passwordHash == "" {
		http.Error(w, `{"error":"login not configured"}`, http.StatusInternalServerError)
		return
	}

	if !CheckPasswordHash(req.Password, h.passwordHash) {
		http.Error(w, `{"error":"wrong password"}`, http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken("admin")
	if err != nil {
		http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token, Message: "login ok"})
}
