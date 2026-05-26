package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kapibara-pro/sway-project/backend/internal/domain"
	"github.com/kapibara-pro/sway-project/backend/internal/llm"
	"github.com/kapibara-pro/sway-project/backend/internal/store"
)

const (
	CodeInvalidInput       = "INVALID_INPUT"
	CodeTextTooLong        = "TEXT_TOO_LONG"
	CodeFullAccessRequired = "FULL_ACCESS_REQUIRED"
	CodeSafetyBlocked      = "SAFETY_BLOCKED"
	CodeModelTimeout       = "MODEL_TIMEOUT"
	CodeRateLimited        = "RATE_LIMITED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

type Server struct {
	Store      store.Store
	LLM        *llm.Client
	Now        func() time.Time
	Logger     *log.Logger
	HistoryTTL time.Duration
}

func NewServer(st store.Store, client *llm.Client) *Server {
	return &Server{Store: st, LLM: client, Now: func() time.Time { return time.Now().UTC() }, HistoryTTL: 30 * 24 * time.Hour}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /api/v1/providers/status", s.handleProviderStatus)
	mux.HandleFunc("PUT /api/v1/providers/config", s.handleProviderConfig)
	mux.HandleFunc("POST /api/v1/providers/test", s.handleProviderTest)
	mux.HandleFunc("DELETE /api/v1/providers/config", s.handleProviderDelete)
	mux.HandleFunc("POST /api/v1/chat-assist/generate", s.handleGenerate)
	mux.HandleFunc("GET /api/v1/chat-assist/requests", s.handleRequestLogs)
	return requestIDMiddleware(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleProviderStatus(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := requireDeviceID(w, r)
	if !ok {
		return
	}
	cfg, err := s.Store.GetProviderConfig(r.Context(), deviceID)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]any{"configured": false})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeServiceUnavailable, "provider config unavailable", "请稍后重试", requestIDFromContext(r.Context()), nil)
		return
	}
	writeJSON(w, http.StatusOK, providerStatus(cfg))
}

func (s *Server) handleProviderConfig(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := requireDeviceID(w, r)
	if !ok {
		return
	}
	var req struct {
		Provider string `json:"provider"`
		BaseURL  string `json:"base_url"`
		Model    string `json:"model"`
		APIKey   string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidInput, "invalid json", "请检查模型配置格式", requestIDFromContext(r.Context()), nil)
		return
	}
	provider := strings.TrimSpace(req.Provider)
	if strings.TrimSpace(req.Provider) == "" || strings.TrimSpace(req.Model) == "" {
		writeError(w, http.StatusBadRequest, CodeInvalidInput, "missing provider fields", "请完整填写 Provider、Base URL、Model 和 API Key", requestIDFromContext(r.Context()), nil)
		return
	}
	if !strings.EqualFold(provider, "mock") && (strings.TrimSpace(req.BaseURL) == "" || strings.TrimSpace(req.APIKey) == "") {
		writeError(w, http.StatusBadRequest, CodeInvalidInput, "missing provider fields", "请完整填写 Provider、Base URL、Model 和 API Key", requestIDFromContext(r.Context()), nil)
		return
	}
	cfg := domain.ProviderConfig{
		DeviceID:     deviceID,
		Provider:     provider,
		BaseURL:      strings.TrimRight(strings.TrimSpace(req.BaseURL), "/"),
		Model:        strings.TrimSpace(req.Model),
		EncryptedKey: strings.TrimSpace(req.APIKey),
		KeyMask:      maskAPIKey(req.APIKey),
		UpdatedAt:    s.now(),
	}
	if err := s.Store.SaveProviderConfig(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, CodeServiceUnavailable, "save provider failed", "保存模型配置失败，请稍后重试", requestIDFromContext(r.Context()), nil)
		return
	}
	writeJSON(w, http.StatusOK, providerStatus(cfg))
}

func (s *Server) handleProviderTest(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := requireDeviceID(w, r)
	if !ok {
		return
	}
	cfg, err := s.Store.GetProviderConfig(r.Context(), deviceID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, CodeInvalidInput, "provider not configured", "请先在 App 中配置模型", requestIDFromContext(r.Context()), nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeServiceUnavailable, "provider config unavailable", "请稍后重试", requestIDFromContext(r.Context()), nil)
		return
	}
	cfg.LastTestedAt = s.now()
	if err := s.LLM.TestProvider(r.Context(), cfg); err != nil {
		cfg.LastTestOK = false
		cfg.LastTestError = sanitizeErr(err)
		_ = s.Store.SaveProviderConfig(r.Context(), cfg)
		writeError(w, http.StatusBadGateway, CodeServiceUnavailable, "provider test failed", "模型连通性测试失败，请检查配置", requestIDFromContext(r.Context()), map[string]any{"provider_error": cfg.LastTestError})
		return
	}
	cfg.LastTestOK = true
	cfg.LastTestError = ""
	_ = s.Store.SaveProviderConfig(r.Context(), cfg)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": providerStatus(cfg)})
}

func (s *Server) handleProviderDelete(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := requireDeviceID(w, r)
	if !ok {
		return
	}
	if err := s.Store.DeleteProviderConfig(r.Context(), deviceID); err != nil {
		writeError(w, http.StatusInternalServerError, CodeServiceUnavailable, "delete provider failed", "清除模型配置失败，请稍后重试", requestIDFromContext(r.Context()), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	started := s.now()
	requestID := requestIDFromContext(r.Context())
	deviceID, ok := requireDeviceID(w, r)
	if !ok {
		return
	}
	var req domain.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidInput, "invalid json", "请检查输入内容", requestID, nil)
		return
	}
	req = normalizeGenerateRequest(req)
	if validationErr := validateGenerateRequest(req); validationErr != nil {
		writeError(w, validationErr.HTTPStatus, validationErr.Code, validationErr.Message, validationErr.Fallback, requestID, validationErr.Details)
		return
	}
	if blocked, reason := safetyBlocked(req); blocked {
		writeError(w, http.StatusBadRequest, CodeSafetyBlocked, reason, "这段内容不适合直接生成，我可以帮你改成更健康、尊重边界的表达。", requestID, nil)
		return
	}
	cfg, err := s.Store.GetProviderConfig(r.Context(), deviceID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusBadRequest, CodeInvalidInput, "provider not configured", "请先在 App 中配置模型", requestID, nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeServiceUnavailable, "provider config unavailable", "请稍后重试", requestID, nil)
		return
	}
	prompt := llm.BuildPrompt(req)
	out, err := s.LLM.Generate(r.Context(), llm.GenerateInput{ProviderConfig: cfg, Request: req, Prompt: prompt})
	latency := s.now().Sub(started).Milliseconds()
	if err != nil {
		code := CodeServiceUnavailable
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
			code = CodeModelTimeout
		}
		s.saveRequestLog(r.Context(), requestID, deviceID, req, cfg, "failed", code, nil, out.Usage, latency)
		writeError(w, http.StatusBadGateway, code, sanitizeErr(err), "生成暂时不可用，请稍后重试或在 App 内使用本地模板。", requestID, nil)
		return
	}
	out.Usage.LatencyMS = latency
	resp := domain.GenerateResponse{
		RequestID:  requestID,
		Candidates: out.Candidates,
		Safety:     domain.SafetyResult{Blocked: false},
		Usage:      out.Usage,
	}
	s.saveRequestLog(r.Context(), requestID, deviceID, req, cfg, "success", "", out.Candidates, out.Usage, latency)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRequestLogs(w http.ResponseWriter, r *http.Request) {
	deviceID, _ := deviceIDFromRequest(r)
	logs, err := s.Store.ListRequestLogs(r.Context(), deviceID, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, CodeServiceUnavailable, "logs unavailable", "请求记录暂时不可用", requestIDFromContext(r.Context()), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": logs})
}

func (s *Server) saveRequestLog(ctx context.Context, requestID, deviceID string, req domain.GenerateRequest, cfg domain.ProviderConfig, status, code string, candidates []domain.Candidate, usage domain.Usage, latency int64) {
	logItem := domain.RequestLog{
		RequestID:   requestID,
		DeviceID:    deviceID,
		Mode:        req.Mode,
		Source:      req.Source,
		InputPolicy: req.InputPolicy,
		Language:    req.Language,
		Tone:        req.Tone,
		Provider:    cfg.Provider,
		Model:       cfg.Model,
		Status:      status,
		ErrorCode:   code,
		LatencyMS:   latency,
		Usage:       usage,
		PeerLen:     utf8.RuneCountInString(req.PeerMessage),
		DraftLen:    utf8.RuneCountInString(req.Draft),
		ClientCtx:   req.ClientContext,
		CreatedAt:   s.now(),
		ExpiresAt:   s.now().Add(s.HistoryTTL),
	}
	if req.InputPolicy == domain.InputPolicyStoreAllowed {
		logItem.PeerMessage = req.PeerMessage
		logItem.Draft = req.Draft
		logItem.Candidates = candidates
	}
	_ = s.Store.SaveRequestLog(ctx, logItem)
}

func validateGenerateRequest(req domain.GenerateRequest) *validationError {
	if !validIn(req.Mode, []string{domain.ModeRewrite, domain.ModeReply, domain.ModeOpener, domain.ModeComfort, domain.ModeApologize, domain.ModeReject}) {
		return invalidField("mode", "unsupported mode")
	}
	if !validIn(req.Source, []string{domain.SourceApp, domain.SourceKeyboard}) {
		return invalidField("source", "unsupported source")
	}
	if !validIn(req.InputPolicy, []string{domain.InputPolicyEphemeral, domain.InputPolicyStoreAllowed}) {
		return invalidField("input_policy", "unsupported input policy")
	}
	if !validIn(req.Language, []string{"zh-CN", "en-US"}) {
		return invalidField("language", "unsupported language")
	}
	if req.Count <= 0 || req.Count > 3 {
		return invalidField("count", "count must be 1-3")
	}
	if strings.TrimSpace(req.PeerMessage) == "" && strings.TrimSpace(req.Draft) == "" {
		return invalidField("text", "peer_message or draft is required")
	}
	peerLen := utf8.RuneCountInString(req.PeerMessage)
	draftLen := utf8.RuneCountInString(req.Draft)
	if peerLen > 1000 || draftLen > 500 || peerLen+draftLen > 1500 {
		return &validationError{HTTPStatus: http.StatusBadRequest, Code: CodeTextTooLong, Message: "input text too long", Fallback: "内容太长了，请缩短后再生成。", Details: map[string]any{"peer_len": peerLen, "draft_len": draftLen}}
	}
	return nil
}

func normalizeGenerateRequest(req domain.GenerateRequest) domain.GenerateRequest {
	if req.Source == "" {
		req.Source = domain.SourceApp
	}
	if req.InputPolicy == "" {
		req.InputPolicy = domain.InputPolicyEphemeral
	}
	if req.Language == "" {
		req.Language = "zh-CN"
	}
	if req.Tone == "" {
		req.Tone = "gentle"
	}
	if req.Length == "" {
		req.Length = "short"
	}
	if req.RelationshipStage == "" {
		req.RelationshipStage = "unknown"
	}
	if req.Count == 0 {
		req.Count = 3
	}
	return req
}

type validationError struct {
	HTTPStatus int
	Code       string
	Message    string
	Fallback   string
	Details    map[string]any
}

func invalidField(field, reason string) *validationError {
	return &validationError{HTTPStatus: http.StatusBadRequest, Code: CodeInvalidInput, Message: reason, Fallback: "请检查输入后再试。", Details: map[string]any{"field": field, "reason": reason}}
}

func safetyBlocked(req domain.GenerateRequest) (bool, string) {
	text := strings.ToLower(req.PeerMessage + " " + req.Draft)
	blocked := []string{
		"未成年", "未成年人", "自杀", "威胁", "跟踪", "骚扰",
		"minor", "underage", "suicide", "self-harm", "self harm", "threaten", "stalk", "harass",
	}
	for _, token := range blocked {
		if strings.Contains(text, token) {
			return true, "content safety policy matched"
		}
	}
	return false, ""
}

func providerStatus(cfg domain.ProviderConfig) domain.ProviderStatus {
	return domain.ProviderStatus{Configured: true, Provider: cfg.Provider, BaseURL: cfg.BaseURL, Model: cfg.Model, KeyMask: cfg.KeyMask, UpdatedAt: cfg.UpdatedAt, LastTestedAt: cfg.LastTestedAt, LastTestOK: cfg.LastTestOK, LastTestError: cfg.LastTestError}
}

func (s *Server) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

func validIn(v string, allowed []string) bool {
	for _, item := range allowed {
		if v == item {
			return true
		}
	}
	return false
}

func maskAPIKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func sanitizeErr(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 300 {
		msg = msg[:300]
	}
	return msg
}

func requireDeviceID(w http.ResponseWriter, r *http.Request) (string, bool) {
	deviceID, ok := deviceIDFromRequest(r)
	if !ok {
		writeError(w, http.StatusBadRequest, CodeInvalidInput, "missing device id", "缺少设备标识，请重启 App 后再试。", requestIDFromContext(r.Context()), map[string]any{"field": "X-Sway-Device-ID"})
		return "", false
	}
	return deviceID, true
}

func deviceIDFromRequest(r *http.Request) (string, bool) {
	deviceID := strings.TrimSpace(r.Header.Get("X-Sway-Device-ID"))
	if deviceID == "" {
		deviceID = strings.TrimSpace(r.URL.Query().Get("device_id"))
	}
	return deviceID, deviceID != ""
}

func writeError(w http.ResponseWriter, status int, code, message, fallback, requestID string, details map[string]any) {
	writeJSONStatus(w, status, domain.APIError{Code: code, Message: message, FallbackSuggestion: fallback, RequestID: requestID, Details: details})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	writeJSONStatus(w, status, v)
}

func writeJSONStatus(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type ctxKey string

const requestIDKey ctxKey = "request_id"

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = newID("req")
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func newID(prefix string) string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return prefix + "_" + time.Now().UTC().Format("20060102150405")
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}
