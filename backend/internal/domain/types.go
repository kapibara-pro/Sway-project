package domain

import "time"

const (
	ModeRewrite   = "rewrite"
	ModeReply     = "reply"
	ModeOpener    = "opener"
	ModeComfort   = "comfort"
	ModeApologize = "apologize"
	ModeReject    = "reject"

	SourceKeyboard = "keyboard"
	SourceApp      = "app"

	InputPolicyEphemeral    = "ephemeral"
	InputPolicyStoreAllowed = "store_allowed"
)

type ProviderConfig struct {
	DeviceID      string    `json:"device_id"`
	Provider      string    `json:"provider"`
	BaseURL       string    `json:"base_url"`
	Model         string    `json:"model"`
	EncryptedKey  string    `json:"-"`
	KeyMask       string    `json:"key_mask"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastTestedAt  time.Time `json:"last_tested_at,omitempty"`
	LastTestOK    bool      `json:"last_test_ok"`
	LastTestError string    `json:"last_test_error,omitempty"`
}

type ProviderStatus struct {
	Configured    bool      `json:"configured"`
	Provider      string    `json:"provider,omitempty"`
	BaseURL       string    `json:"base_url,omitempty"`
	Model         string    `json:"model,omitempty"`
	KeyMask       string    `json:"key_mask,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	LastTestedAt  time.Time `json:"last_tested_at,omitempty"`
	LastTestOK    bool      `json:"last_test_ok"`
	LastTestError string    `json:"last_test_error,omitempty"`
}

type Candidate struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	ToneLabel     string `json:"tone_label"`
	ScenarioLabel string `json:"scenario_label"`
	RiskLevel     string `json:"risk_level"`
	WhyThisWorks  string `json:"why_this_works,omitempty"`
}

type Usage struct {
	Provider         string `json:"provider,omitempty"`
	Model            string `json:"model,omitempty"`
	PromptTokens     int    `json:"prompt_tokens,omitempty"`
	CompletionTokens int    `json:"completion_tokens,omitempty"`
	TotalTokens      int    `json:"total_tokens,omitempty"`
	LatencyMS        int64  `json:"latency_ms"`
}

type SafetyResult struct {
	Blocked bool   `json:"blocked"`
	Reason  string `json:"reason,omitempty"`
}

type GenerateRequest struct {
	Mode              string            `json:"mode"`
	Source            string            `json:"source"`
	InputPolicy       string            `json:"input_policy"`
	PeerMessage       string            `json:"peer_message"`
	Draft             string            `json:"draft"`
	Tone              string            `json:"tone"`
	RelationshipStage string            `json:"relationship_stage"`
	Language          string            `json:"language"`
	Length            string            `json:"length"`
	Count             int               `json:"count"`
	ClientContext     map[string]string `json:"client_context"`
}

type GenerateResponse struct {
	RequestID          string       `json:"request_id"`
	Candidates         []Candidate  `json:"candidates"`
	Safety             SafetyResult `json:"safety"`
	Usage              Usage        `json:"usage"`
	FallbackSuggestion string       `json:"fallback_suggestion,omitempty"`
}

type APIError struct {
	Code               string         `json:"code"`
	Message            string         `json:"message"`
	FallbackSuggestion string         `json:"fallback_suggestion,omitempty"`
	RequestID          string         `json:"request_id,omitempty"`
	Details            map[string]any `json:"details,omitempty"`
}

type RequestLog struct {
	RequestID   string            `json:"request_id"`
	DeviceID    string            `json:"device_id"`
	Mode        string            `json:"mode"`
	Source      string            `json:"source"`
	InputPolicy string            `json:"input_policy"`
	Language    string            `json:"language"`
	Tone        string            `json:"tone"`
	Provider    string            `json:"provider"`
	Model       string            `json:"model"`
	Status      string            `json:"status"`
	ErrorCode   string            `json:"error_code,omitempty"`
	LatencyMS   int64             `json:"latency_ms"`
	Usage       Usage             `json:"usage"`
	PeerLen     int               `json:"peer_len"`
	DraftLen    int               `json:"draft_len"`
	PeerMessage string            `json:"peer_message,omitempty"`
	Draft       string            `json:"draft,omitempty"`
	Candidates  []Candidate       `json:"candidates,omitempty"`
	ClientCtx   map[string]string `json:"client_context,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
}
