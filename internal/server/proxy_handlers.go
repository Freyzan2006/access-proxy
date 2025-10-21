package server

import (
	"net/http"

	"github.com/Freyzan2006/go-logger-lib/pkg/logger"
)

type requestProcessor struct {
	log logger.Logger
}

func newRequestProcessor(log logger.Logger) *requestProcessor {
	return &requestProcessor{log: log}
}

func (p *requestProcessor) process(req *http.Request) {
	p.setStandardHeaders(req)
	p.removeProblematicHeaders(req)
	p.logRequest(req)
}

func (p *requestProcessor) setStandardHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Access-Proxy-Server/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "identity")
}

func (p *requestProcessor) removeProblematicHeaders(req *http.Request) {
	req.Header.Del("Accept-Encoding")
	req.Header.Del("X-Forwarded-Proto")
}

func (p *requestProcessor) logRequest(req *http.Request) {
	p.log.Infof("➡️  Forwarding to %s %s", req.Method, req.URL.String())
}

type responseProcessor struct {
	log logger.Logger
}

func newResponseProcessor(log logger.Logger) *responseProcessor {
	return &responseProcessor{log: log}
}

func (p *responseProcessor) logResponse(resp *http.Response) {
	p.log.Infof("📨 Response: %d %s for %s", resp.StatusCode, resp.Status, resp.Request.URL.Path)
}

type errorHandler struct {
	log logger.Logger
}

func newErrorHandler(log logger.Logger) *errorHandler {
	return &errorHandler{log: log}
}

func (h *errorHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Errorf("❌ Proxy error: %v", err)
	h.log.Errorf("❌ Request: %s %s", r.Method, r.URL.String())
	
	h.writeErrorResponse(w, err)
}

func (h *errorHandler) writeErrorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadGateway)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"error": "Bad Gateway", "message": "` + err.Error() + `"}`))
}