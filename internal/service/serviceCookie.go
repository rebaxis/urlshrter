package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTCookieConfig struct {
	SecretKey       string
	CookieName      string
	TokenExpiration time.Time
	Secure          bool
	HTTPOnly        bool
	SameSite        http.SameSite
	MaxAge          int
}

type JWTCookieService struct {
	config JWTCookieConfig
}

var (
	ErrInvalidToken = fmt.Errorf("invalid token")
)

func NewJWTCookieService(config JWTCookieConfig) *JWTCookieService {
	if config.CookieName == "" {
		config.CookieName = "jwt_token"
	}
	if config.TokenExpiration.IsZero() {
		config.TokenExpiration = time.Now().AddDate(100, 0, 0)
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}
	config.Secure = true
	config.HTTPOnly = true
	config.MaxAge = 3600

	return &JWTCookieService{
		config: config,
	}
}

// GenerateToken создает JWT токен
func (j *JWTCookieService) generateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		// UserData: userData,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(j.config.TokenExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "jwt-cookie-manager",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// ParseToken парсит и валидирует JWT токен
func (j *JWTCookieService) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// SetTokenCookie устанавливает JWT токен в cookie
func (j *JWTCookieService) SetJWTCookie(w http.ResponseWriter, userID string) error {
	token, err := j.generateToken(userID)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     j.config.CookieName,
		Value:    token,
		Expires:  j.config.TokenExpiration,
		Secure:   j.config.Secure,
		HttpOnly: j.config.HTTPOnly,
		SameSite: j.config.SameSite,
		MaxAge:   j.config.MaxAge,
	})

	return nil
}

// GetClaimsFromRequest извлекает claims из cookie запроса
func (j *JWTCookieService) GetClaimsFromRequest(r *http.Request) (*Claims, error) {
	cookie, err := r.Cookie(j.config.CookieName)
	if err != nil {
		return nil, err
	}

	return j.parseToken(cookie.Value)
}
