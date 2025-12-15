package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Bessima/diplom-gomarket/internal/handlers/schemas"
	"github.com/Bessima/diplom-gomarket/internal/middlewares/logger"
	"github.com/Bessima/diplom-gomarket/internal/repository"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/Bessima/diplom-gomarket/internal/models"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AuthHandler struct {
	jwtConfig   *JWTConfig
	UserStorage repository.UserStorageRepositoryI
}

func NewAuthHandler(jwtConfig *JWTConfig, storage repository.UserStorageRepositoryI) *AuthHandler {
	return &AuthHandler{
		jwtConfig:   jwtConfig,
		UserStorage: storage,
	}
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req schemas.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 50 {
		http.Error(w, "Username must be between 3 and 50 characters", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	existingUser, _ := h.UserStorage.GetUserByUsername(req.Username)
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		logger.Log.Error(fmt.Errorf("error generate password hash %v", err).Error())
		return
	}

	user, err := h.UserStorage.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		logger.Log.Error(fmt.Errorf("error creating user in DB: %v", err).Error())
		return
	}

	// Автоматическая аутентификация после регистрации
	accessToken, refreshToken, err := h.generateTokens(user)
	if err != nil {
		http.Error(w, "Error generating tokens", http.StatusInternalServerError)
		return
	}

	// Установка токенов в cookies (можно также использовать Authorization header)
	h.setTokensInCookies(w, accessToken, refreshToken)

	// Отправка ответа
	response := schemas.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    time.Now().Add(h.jwtConfig.AccessTokenTTL).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req schemas.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.UserStorage.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, err := h.generateTokens(user)
	if err != nil {
		http.Error(w, "Error generating tokens", http.StatusInternalServerError)
		return
	}

	h.setTokensInCookies(w, accessToken, refreshToken)

	response := schemas.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    time.Now().Add(h.jwtConfig.AccessTokenTTL).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Очистка cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			http.Error(w, "Refresh token required", http.StatusUnauthorized)
			return
		}
		refreshToken := authHeader[7:]

		claims, err := h.ValidateToken(refreshToken)
		if err != nil {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		user, err := h.UserStorage.GetUserByID(claims.UserID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		accessToken, newRefreshToken, err := h.generateTokens(user)
		if err != nil {
			http.Error(w, "Error generating tokens", http.StatusInternalServerError)
			return
		}

		h.setTokensInCookies(w, accessToken, newRefreshToken)

		response := schemas.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    time.Now().Add(h.jwtConfig.AccessTokenTTL).Unix(),
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	claims, err := h.ValidateToken(refreshCookie.Value)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	user, err := h.UserStorage.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	accessToken, newRefreshToken, err := h.generateTokens(user)
	if err != nil {
		http.Error(w, "Error generating tokens", http.StatusInternalServerError)
		return
	}

	h.setTokensInCookies(w, accessToken, newRefreshToken)

	response := schemas.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    time.Now().Add(h.jwtConfig.AccessTokenTTL).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) generateTokens(user *models.User) (string, string, error) {
	// Access token
	accessClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.jwtConfig.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}
	println(accessClaims.UserID)
	println("accessClaims")

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(h.jwtConfig.SecretKey))
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshClaims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.jwtConfig.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.jwtConfig.SecretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (h *AuthHandler) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	println("VALIDATE")

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	println(claims)
	println("66666")

	return claims, nil
}

func (h *AuthHandler) setTokensInCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	// Access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(h.jwtConfig.AccessTokenTTL),
		HttpOnly: true,
		Secure:   false, // Только для HTTPS в production
		SameSite: http.SameSiteStrictMode,
	})

	// Refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(h.jwtConfig.RefreshTokenTTL),
		HttpOnly: true,
		Secure:   false, // Только для HTTPS в production
		SameSite: http.SameSiteStrictMode,
	})
}

// GetUserFromContext извлекает пользователя из контекста
func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}
