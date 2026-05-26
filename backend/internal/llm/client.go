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
			{"role": "system", "content": BuildSystemPrompt(input.Request)},
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
	if len(payload.Candidates) < limit {
		return nil, fmt.Errorf("provider returned %d candidates, want %d", len(payload.Candidates), limit)
	}
	if len(payload.Candidates) > limit {
		payload.Candidates = payload.Candidates[:limit]
	}
	for i := range payload.Candidates {
		payload.Candidates[i].Text = strings.TrimSpace(payload.Candidates[i].Text)
		if payload.Candidates[i].Text == "" {
			return nil, fmt.Errorf("candidate %d has empty text", i+1)
		}
		if payload.Candidates[i].ID == "" {
			payload.Candidates[i].ID = fmt.Sprintf("c%d", i+1)
		}
		if payload.Candidates[i].ScenarioLabel == "" || payload.Candidates[i].ScenarioLabel != req.Mode {
			payload.Candidates[i].ScenarioLabel = req.Mode
		}
		if payload.Candidates[i].ToneLabel == "" {
			payload.Candidates[i].ToneLabel = req.Tone
		}
		if payload.Candidates[i].RiskLevel == "" {
			payload.Candidates[i].RiskLevel = "low"
		}
		if payload.Candidates[i].RiskLevel != "low" && payload.Candidates[i].RiskLevel != "medium" {
			payload.Candidates[i].RiskLevel = "low"
		}
	}
	return payload.Candidates, nil
}
