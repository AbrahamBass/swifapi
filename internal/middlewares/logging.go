package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/AbrahamBass/swiftapi/internal/responses"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.New().String()

			cwr := responses.NewCustomResponseWriter(w)

			defer func() {
				if err := recover(); err != nil {

					logger.Error("Recovered from panic",
						zap.String("request_id", requestID),
						zap.Any("error", err),
					)

					cwr.WriteHeader(http.StatusInternalServerError)

					if !cwr.Written() {
						errorResponse := map[string]string{
							"error":      "Internal Server Error",
							"request_id": requestID,
						}
						if err := json.NewEncoder(cwr).Encode(errorResponse); err != nil {
							logger.Error("Failed to encode error response",
								zap.String("request_id", requestID),
								zap.Error(err),
							)
						}
					}
				}
			}()

			next.ServeHTTP(cwr, r)

			latency := time.Since(start)

			logger.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", cwr.StatusCode),
				zap.String("status_text", http.StatusText(cwr.StatusCode)),
				zap.Duration("latency", latency),
				zap.String("ip", r.RemoteAddr),
				zap.String("user_agent", r.Header.Get("User-Agent")),
				zap.String("referer", r.Header.Get("Referer")),
				zap.String("protocol", r.Proto),
				zap.String("host", r.Host),
				zap.Int64("request_size", r.ContentLength),
				zap.Int("response_size", cwr.Size),
				zap.String("request_id", requestID),
			)

			if strings.Contains(r.URL.RawQuery, "<script>") {
				logger.Warn("Possible XSS attempt detected",
					zap.String("request_id", requestID),
				)
			}

		})
	}
}
