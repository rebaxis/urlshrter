// Package service содержит сервисы для работы с JWT токенами и cookies.
package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims представляет структуру JWT claims с пользовательскими данными.
// Содержит UserID и стандартные JWT поля (exp, iat, nbf, iss).
type Claims struct {
	UserID string `json:"user_id"` // Идентификатор пользователя
	jwt.RegisteredClaims
}

// JWTCookieConfig содержит конфигурацию для JWT токенов и cookies.
// Определяет параметры безопасности и срок действия токенов.
type JWTCookieConfig struct {
	SecretKey       string        // Секретный ключ для подписи токенов
	CookieName      string        // Имя cookie для хранения токена
	TokenExpiration time.Time     // Время истечения токена
	Secure          bool          // Флаг Secure для cookie (только HTTPS)
	HTTPOnly        bool          // Флаг HttpOnly для защиты от XSS
	SameSite        http.SameSite // Политика SameSite для CSRF защиты
}

// JWTCookieService предоставляет функции для работы с JWT токенами в cookies.
// Управляет созданием, валидацией и извлечением токенов из HTTP запросов.
type JWTCookieService struct {
	config JWTCookieConfig
}

var (
	// ErrInvalidToken возвращается при валидации невалидного JWT токена.
	ErrInvalidToken = fmt.Errorf("invalid token")
)

// NewJWTCookieService создает новый сервис для работы с JWT в cookies.
// Устанавливает значения по умолчанию для незаполненных полей конфигурации.
// По умолчанию: CookieName="jwt_token", Expiration=100 лет, SameSite=Lax, HttpOnly=true, Secure=false.
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
	config.Secure = false
	config.HTTPOnly = true

	return &JWTCookieService{
		config: config,
	}
}

// generateToken создает JWT токен для указанного пользователя.
// Токен содержит UserID и стандартные claims (exp, iat, nbf, iss).
// Подписывается с использованием HMAC SHA256 и секретного ключа из конфигурации.
func (j *JWTCookieService) generateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
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

// parseToken парсит и валидирует JWT токен.
// Проверяет подпись токена с использованием секретного ключа.
// Возвращает ErrInvalidToken если токен невалиден или истек срок действия.
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

// SetJWTCookie создает JWT токен для пользователя и устанавливает его в HTTP cookie.
// Использует настройки из конфигурации (имя cookie, срок действия, флаги безопасности).
// Вызывается при успешной аутентификации или регистрации пользователя.
func (j *JWTCookieService) SetJWTCookie(w *http.ResponseWriter, userID string) error {
	token, err := j.generateToken(userID)
	if err != nil {
		return err
	}

	http.SetCookie(*w, &http.Cookie{
		Name:     j.config.CookieName,
		Value:    token,
		Expires:  j.config.TokenExpiration,
		Secure:   j.config.Secure,
		HttpOnly: j.config.HTTPOnly,
		SameSite: j.config.SameSite,
	})

	return nil
}

// GetClaimsFromRequest извлекает и валидирует JWT claims из cookie HTTP запроса.
// Читает cookie с именем из конфигурации, парсит токен и возвращает claims.
// Используется в middleware для идентификации пользователя.
func (j *JWTCookieService) GetClaimsFromRequest(r *http.Request) (*Claims, error) {
	cookie, err := r.Cookie(j.config.CookieName)
	if err != nil {
		return nil, err
	}

	return j.parseToken(cookie.Value)
}

// ParseToken валидирует JWT токен из строки и возвращает claims.
// Используется в gRPC interceptor для извлечения userID из метаданных.
func (j *JWTCookieService) ParseToken(tokenString string) (*Claims, error) {
	return j.parseToken(tokenString)
}
