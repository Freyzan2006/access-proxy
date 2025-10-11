// internal/middleware/logging.go
package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// LoggingMiddleware логирует все входящие запросы и ответы
func LoggingMiddleware(log logger.Logger, enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			
			// Логируем входящий запрос
			log.Infof("📥 INCOMING REQUEST: %s %s %s", r.Method, r.URL.Path, r.Proto)
			log.Infof("📍 Remote Addr: %s", r.RemoteAddr)
			log.Infof("🌐 User Agent: %s", r.UserAgent())
			log.Infof("📨 Headers: %v", r.Header)

			// Читаем тело запроса если есть
			var requestBody bytes.Buffer
			if r.Body != nil {
				tee := io.TeeReader(r.Body, &requestBody)
				bodyBytes, _ := io.ReadAll(tee)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				
				if len(bodyBytes) > 0 {
					log.Infof("📦 Request Body: %s", string(bodyBytes))
				}
			}

			// Создаем recorder для перехвата ответа
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Обрабатываем запрос
			next.ServeHTTP(recorder, r)

			// Логируем ответ
			duration := time.Since(start)
			log.Infof("📤 RESPONSE: %d %s", recorder.statusCode, http.StatusText(recorder.statusCode))
			log.Infof("⏱️  Duration: %v", duration)
			log.Infof("📊 Response Size: %d bytes", recorder.body.Len())
			
			if recorder.body.Len() > 0 && recorder.body.Len() < 1024 { // Логируем только маленькие тела
				log.Infof("📦 Response Body: %s", recorder.body.String())
			}
			
			log.Infof("🔚 Request completed: %s %s", r.Method, r.URL.Path)
		})
	}
}