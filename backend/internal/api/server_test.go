package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kapibara-pro/sway-project/backend/internal/domain"
	"github.com/kapibara-pro/sway-project/backend/internal/llm"
	"github.com/kapibara-pro/sway-project/backend/internal/store"
)

func TestProviderStatusEmpty(t *testing.T) {
	s := NewServer(store.NewMemoryStore(), llm.NewClient(time.Second))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/providers/status", nil)
	req.Header.Set("X-Sway-Device-ID", "device-1")
	rr := httptest.NewRecorder()
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["configured"] != false {
		t.Fatalf("configured=%v, want false", got["configured"])
	}
}

func TestProviderConfigMasksAPIKey(t *testing.T) {
	s := NewServer(store.NewMemoryStore(), llm.NewClient(time.Second))
	body := bytes.NewBufferString(`{"provider":"openai","base_url":"https://api.example.com/v1","model":"gpt-test","api_key":"sk-1234567890"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/providers/config", body)
	req.Header.Set("X-Sway-Device-ID", "device-1")
	rr := httptest.NewRecorder()
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if bytes.Contains(rr.Body.Bytes(), []byte("sk-1234567890")) {
		t.Fatalf("response leaked api key: %s", rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("sk-1****7890")) {
		t.Fatalf("response missing key mask: %s", rr.Body.String())
	}
}

func TestGenerateValidatesTextLength(t *testing.T) {
	s := NewServer(store.NewMemoryStore(), llm.NewClient(time.Second))
	longText := make([]rune, 1001)
	for i := range longText {
		longText[i] = '你'
	}
	body := bytes.NewBufferString(`{"mode":"reply","source":"app","input_policy":"ephemeral","peer_message":"` + string(longText) + `","language":"zh-CN","count":3}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat-assist/generate", body)
	req.Header.Set("X-Sway-Device-ID", "device-1")
	rr := httptest.NewRecorder()
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(CodeTextTooLong)) {
		t.Fatalf("body missing code: %s", rr.Body.String())
	}
}

func TestGenerateRequiresProvider(t *testing.T) {
	s := NewServer(store.NewMemoryStore(), llm.NewClient(time.Second))
	body := bytes.NewBufferString(`{"mode":"reply","source":"app","input_policy":"ephemeral","peer_message":"你好","language":"zh-CN","count":3}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat-assist/generate", body)
	req.Header.Set("X-Sway-Device-ID", "device-1")
	rr := httptest.NewRecorder()
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(CodeInvalidInput)) {
		t.Fatalf("body missing code: %s", rr.Body.String())
	}
}

func TestGenerateWithMockProvider(t *testing.T) {
	st := store.NewMemoryStore()
	s := NewServer(st, llm.NewClient(time.Second))
	if err := st.SaveProviderConfig(context.Background(), domain.ProviderConfig{
		DeviceID:     "device-1",
		Provider:     "mock",
		Model:        "mock-chat",
		EncryptedKey: "",
		KeyMask:      "****",
	}); err != nil {
		t.Fatal(err)
	}
	body := bytes.NewBufferString(`{"mode":"reply","source":"app","input_policy":"store_allowed","peer_message":"你今天怎么都不找我聊天？","language":"zh-CN","count":3}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat-assist/generate", body)
	req.Header.Set("X-Sway-Device-ID", "device-1")
	rr := httptest.NewRecorder()
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var got domain.GenerateResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Candidates) != 3 {
		t.Fatalf("candidates=%d, want 3", len(got.Candidates))
	}
	if got.Usage.Provider != "mock" {
		t.Fatalf("provider=%s, want mock", got.Usage.Provider)
	}
}

func TestSaveRequestLogEphemeralDoesNotStoreRawText(t *testing.T) {
	st := store.NewMemoryStore()
	s := NewServer(st, llm.NewClient(time.Second))
	s.saveRequestLog(context.Background(), "req_1", "device-1", minimalGenerateRequest("ephemeral"), testProvider(), "success", "", nil, domainUsage(), 10)
	logs, err := st.ListRequestLogs(context.Background(), "device-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs=%d", len(logs))
	}
	if logs[0].PeerMessage != "" || logs[0].Draft != "" {
		t.Fatalf("raw text stored for ephemeral: %+v", logs[0])
	}
}

func minimalGenerateRequest(policy string) domain.GenerateRequest {
	return domain.GenerateRequest{Mode: "reply", Source: "app", InputPolicy: policy, PeerMessage: "你好", Language: "zh-CN", Count: 3}
}

func testProvider() domain.ProviderConfig {
	return domain.ProviderConfig{Provider: "openai", Model: "test"}
}

func domainUsage() domain.Usage {
	return domain.Usage{LatencyMS: 10}
}
