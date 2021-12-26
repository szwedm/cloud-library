package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/szwedm/cloud-library/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type authentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type token struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

type authHandler struct {
	storage storage.Users
}

func newAuthHandler(u storage.Users) *authHandler {
	return &authHandler{
		storage: u,
	}
}

func (h *authHandler) signin(w http.ResponseWriter, r *http.Request) {
	var authDetails authentication
	if err := json.NewDecoder(r.Body).Decode(&authDetails); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err)
		r.Body.Close()
		return
	}
	defer r.Body.Close()

	dto, err := h.storage.GetUserByUsername(authDetails.Username)
	if err != nil {
		if _, ok := err.(*storage.UserNotFoundErr); ok {
			respondWithError(w, http.StatusNotFound, fmt.Errorf("user %s not found", authDetails.Username))
			return
		}
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(dto.Password), []byte(authDetails.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, err)
		return
	}

	validToken, err := h.generateJWT(dto.Username, dto.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	token := token{
		Username: dto.Username,
		Role:     dto.Role,
		Token:    validToken,
	}

	body, err := json.Marshal(token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, body)
}

func (h *authHandler) generateJWT(username, role string) (string, error) {
	signingKey := []byte(os.Getenv("APP_JWT_SIGN_KEY"))
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = username
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("unable to sign JWT: %w", err)
	}
	return tokenString, nil
}
