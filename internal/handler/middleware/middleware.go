package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

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
		SetJWTCookie(w http.ResponseWriter, userID string) error
	}

	JWTCookieManager interface {
		ClaimsGetter
		JWTCookieSetter
	}
)

func AvtorizationMw(jwtCookieService JWTCookieManager) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := jwtCookieService.GetClaimsFromRequest(r)
			if err == http.ErrNoCookie {
				userID := lib.GenerateRandomAlphabetString(6)
				err = jwtCookieService.SetJWTCookie(w, userID)
				if err != nil {
					http.Error(w, "Some problem: "+err.Error(), http.StatusBadRequest)
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

func BuildMwChain(f http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) == 0 {
		return f
	}

	return m[0](BuildMwChain(f, m[1:]...))
}
