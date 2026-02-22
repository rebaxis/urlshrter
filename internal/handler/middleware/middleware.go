// Package middleware содержит HTTP middleware для обработки запросов.
// Включает middleware для логирования, сжатия, авторизации, валидации и аудита.
package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/rebaxis/urlshrter/internal/audit"
	"github.com/rebaxis/urlshrter/internal/config/logger"
	"github.com/rebaxis/urlshrter/internal/lib"
	"github.com/rebaxis/urlshrter/internal/service"
)

type (
	Middleware func(http.HandlerFunc) http.HandlerFunc

	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	gzipWriter struct {
		http.ResponseWriter
		Writer io.Writer
	}

	auditResponseWriter struct {
		http.ResponseWriter
		status int
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

var LoggingMw = func(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logger.GetLogger()

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	})
}

var CompressMw = func(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Type"), "text/plain") &&
			!strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			h(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h(w, r)
			return
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to create gzip reader "+err.Error(), http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			decompressedBody, err := io.ReadAll(gzipReader)
			if err != nil {
				http.Error(w, "Failed to read decompressed body", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decompressedBody))
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")

		h(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

var ValidatingBodyMw = func(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем что тело запроса не пустое
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) == "" {
			http.Error(w, "Body is empty!", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		h.ServeHTTP(w, r)
	})
}

type (
	ClaimsGetter interface {
		GetClaimsFromRequest(r *http.Request) (*service.Claims, error)
	}

	JWTCookieSetter interface {
		SetJWTCookie(w *http.ResponseWriter, userID string) error
	}

	JWTCookieManager interface {
		ClaimsGetter
		JWTCookieSetter
	}
)

func AuthorizationMw(jwtCookieService JWTCookieManager) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := jwtCookieService.GetClaimsFromRequest(r)
			if err == http.ErrNoCookie || err == service.ErrInvalidToken || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				userID, genErr := lib.GenerateRandomAlphabetString(6)
				if genErr != nil {
					http.Error(w, "Failed to generate user ID: "+genErr.Error(), http.StatusInternalServerError)
					return
				}
				if setErr := jwtCookieService.SetJWTCookie(&w, userID); setErr != nil {
					http.Error(w, "Some problem: "+setErr.Error(), http.StatusBadRequest)
					return
				}
				r.Header.Add("X-User-ID", userID)
				h.ServeHTTP(w, r)
				return
			}
			if err != nil {
				http.Error(w, "Some problem: "+err.Error(), http.StatusBadRequest)
				return
			}

			if claims.UserID != "" {
				r.Header.Add("X-User-ID", claims.UserID)
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// MW для Аудита
func AuditMw(action string, subject *audit.Subject) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var originalURL string
			if action == "shorten" {

				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil && len(bodyBytes) > 0 {
					r.Body.Close()
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					var jsonBody map[string]interface{}
					if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
						if url, ok := jsonBody["url"].(string); ok {
							originalURL = url
						}
					} else {
						originalURL = strings.TrimSpace(string(bodyBytes))
					}
				}
			}

			auditWriter := &auditResponseWriter{
				ResponseWriter: w,
				status:         0,
			}

			h.ServeHTTP(auditWriter, r)

			shouldAudit := false
			if action == "follow" {
				shouldAudit = auditWriter.status >= 200 && auditWriter.status < 400
			} else {
				shouldAudit = auditWriter.status >= 200 && auditWriter.status < 300
			}

			if shouldAudit {
				var url string
				if action == "follow" {
					url = auditWriter.Header().Get("Location")
				} else {
					url = originalURL
				}

				if url != "" {
					event := audit.Event{
						Timestamp: time.Now().Unix(),
						Action:    action,
						UserID:    r.Header.Get("X-User-ID"),
						URL:       url,
					}

					subject.Notify(event)
				}
			}
		})
	}
}

func BuildMwChain(f http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) == 0 {
		return f
	}

	return m[0](BuildMwChain(f, m[1:]...))
}
