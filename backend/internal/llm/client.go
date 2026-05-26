package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kapibara-pro/sway-project/backend/internal/domain"
)

type Client struct {
	HTTPClient *http.Client
	Timeout    time.Duration
}

type GenerateInput struct {
	ProviderConfig domain.ProviderConfig
	Request        domain.GenerateRequest
	Prompt         string
}

type GenerateOutput struct {
	Candidates []domain.Candidate
	Usage      domain.Usage
}

func NewClient(timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &Client{HTTPClient: &http.Client{Timeout: timeout}, Timeout: timeout}
}

func (c *Client) TestProvider(ctx context.Context, cfg domain.ProviderConfig) error {
	if isMockProvider(cfg.Provider) {
		return nil
	}
	_, err := c.Generate(ctx, GenerateInput{
		ProviderConfig: cfg,
		Request: domain.GenerateRequest{
			Mode:        domain.ModeRewrite,
			Source:      domain.SourceApp,
			Language:    "zh-CN",
			Tone:        "concise",
			Length:      "short",
			Count:       1,
			Draft:       "你好，今天辛苦了。",
			InputPolicy: domain.InputPolicyEphemeral,
		},
		Prompt: "请把这句话改得自然一点，只返回一句话。",
	})
	return err
}

func (c *Client) Generate(ctx context.Context, input GenerateInput) (GenerateOutput, error) {
	cfg := input.ProviderConfig
	if isMockProvider(cfg.Provider) {
		return mockGenerate(input.Request, cfg), nil
	}
	apiKey := strings.TrimSpace(cfg.EncryptedKey)
	if apiKey == "" {
		return GenerateOutput{}, fmt.Errorf("missing api key")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return GenerateOutput{}, fmt.Errorf("missing base url")
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		return GenerateOutput{}, fmt.Errorf("missing model")
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt(input.Request)},
			{"role": "user", "content": input.Prompt},
		},
		"temperature":     0.7,
		"response_format": map[string]string{"type": "json_object"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return GenerateOutput{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return GenerateOutput{}, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	hc := c.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return GenerateOutput{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return GenerateOutput{}, fmt.Errorf("provider status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return GenerateOutput{}, fmt.Errorf("decode provider response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return GenerateOutput{}, fmt.Errorf("provider returned no choices")
	}
	candidates, err := parseCandidates(parsed.Choices[0].Message.Content, input.Request)
	if err != nil {
		return GenerateOutput{}, err
	}
	return GenerateOutput{
		Candidates: candidates,
		Usage: domain.Usage{
			Provider:         cfg.Provider,
			Model:            cfg.Model,
			PromptTokens:     parsed.Usage.PromptTokens,
			CompletionTokens: parsed.Usage.CompletionTokens,
			TotalTokens:      parsed.Usage.TotalTokens,
		},
	}, nil
}

func isMockProvider(provider string) bool {
	return strings.EqualFold(strings.TrimSpace(provider), "mock")
}

func mockGenerate(req domain.GenerateRequest, cfg domain.ProviderConfig) GenerateOutput {
	count := req.Count
	if count <= 0 || count > 3 {
		count = 3
	}
	texts := []string{
		"我明白你的意思，也想认真回应你。我们可以慢慢说，不急着把话讲满。",
		"你这么说我有点在意，也有点开心。要不我们换个轻松点的方式聊？",
		"我想把这件事说清楚一点：我在乎你的感受，也希望我们都舒服。",
	}
	if req.Language == "en-US" {
		texts = []string{
			"I get what you mean, and I want to answer this with care. We do not have to rush it.",
			"That actually matters to me. Maybe we can keep it light and talk it through?",
			"I want to say this clearly: I care about how you feel, and I want this to feel comfortable for both of us.",
		}
	}
	candidates := make([]domain.Candidate, 0, count)
	for i := 0; i < count; i++ {
		candidates = append(candidates, domain.Candidate{
			ID:            fmt.Sprintf("c%d", i+1),
			Text:          texts[i],
			ToneLabel:     req.Tone,
			ScenarioLabel: req.Mode,
			RiskLevel:     "low",
			WhyThisWorks:  "保留原意，同时让表达更自然、有分寸。",
		})
	}
	return GenerateOutput{
		Candidates: candidates,
		Usage: domain.Usage{
			Provider:         cfg.Provider,
			Model:            cfg.Model,
			PromptTokens:     120,
			CompletionTokens: 80,
			TotalTokens:      200,
		},
	}
}

func systemPrompt(req domain.GenerateRequest) string {
	lang := "中文"
	if req.Language == "en-US" {
		lang = "English"
	}
	return "你是 Sway/言和的高情商聊天表达助手。只生成用户可自行选择发送的候选文案，不操控、不骚扰、不自动发送。必须输出严格 JSON：{\"candidates\":[{\"text\":\"...\",\"tone_label\":\"...\",\"scenario_label\":\"...\",\"risk_level\":\"low\",\"why_this_works\":\"...\"}]}。语言：" + lang
}

func BuildPrompt(req domain.GenerateRequest) string {
	return fmt.Sprintf(`模式: %s
语气: %s
关系阶段: %s
长度: %s
语言: %s
对方消息:
%s

我的草稿:
%s

请返回 %d 条候选，每条尽量不超过 80 字。`, req.Mode, req.Tone, req.RelationshipStage, req.Length, req.Language, req.PeerMessage, req.Draft, req.Count)
}

func parseCandidates(content string, req domain.GenerateRequest) ([]domain.Candidate, error) {
	var payload struct {
		Candidates []domain.Candidate `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return nil, fmt.Errorf("decode candidates json: %w", err)
	}
	if len(payload.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in provider response")
	}
	limit := req.Count
	if limit <= 0 || limit > 3 {
		limit = 3
	}
	if len(payload.Candidates) > limit {
		payload.Candidates = payload.Candidates[:limit]
	}
	for i := range payload.Candidates {
		if payload.Candidates[i].ID == "" {
			payload.Candidates[i].ID = fmt.Sprintf("c%d", i+1)
		}
		if payload.Candidates[i].RiskLevel == "" {
			payload.Candidates[i].RiskLevel = "low"
		}
	}
	return payload.Candidates, nil
}
