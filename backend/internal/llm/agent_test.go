package llm

import (
	"strings"
	"testing"

	"github.com/kapibara-pro/sway-project/backend/internal/domain"
)

func TestBuildPromptHasModeSpecificGuidance(t *testing.T) {
	tests := []struct {
		mode string
		want string
	}{
		{domain.ModeRewrite, "不改变用户原意"},
		{domain.ModeReply, "对方消息的情绪"},
		{domain.ModeOpener, "轻松"},
		{domain.ModeComfort, "避免说教"},
		{domain.ModeApologize, "不甩锅"},
		{domain.ModeReject, "清楚表达边界"},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			prompt := BuildPrompt(testGenerateRequest(tt.mode, "zh-CN"))
			if !strings.Contains(prompt, tt.want) {
				t.Fatalf("prompt for %s missing %q:\n%s", tt.mode, tt.want, prompt)
			}
			if !strings.Contains(prompt, `scenario_label 固定为 "`+tt.mode+`"`) {
				t.Fatalf("prompt for %s missing fixed scenario label", tt.mode)
			}
			if !strings.Contains(prompt, "只能返回 JSON") {
				t.Fatalf("prompt for %s missing json contract", tt.mode)
			}
		})
	}
}

func TestBuildPromptSupportsEnglish(t *testing.T) {
	prompt := BuildPrompt(testGenerateRequest(domain.ModeReject, "en-US"))
	if !strings.Contains(prompt, "State the boundary clearly") {
		t.Fatalf("english prompt missing english mode goal:\n%s", prompt)
	}
	system := BuildSystemPrompt(testGenerateRequest(domain.ModeReply, "en-US"))
	if !strings.Contains(system, "English") {
		t.Fatalf("system prompt missing english language:\n%s", system)
	}
}

func TestMockGenerateReturnsModeSpecificCandidates(t *testing.T) {
	modes := []string{
		domain.ModeRewrite,
		domain.ModeReply,
		domain.ModeOpener,
		domain.ModeComfort,
		domain.ModeApologize,
		domain.ModeReject,
	}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			out := mockGenerate(testGenerateRequest(mode, "zh-CN"), domain.ProviderConfig{Provider: "mock", Model: "mock-chat"})
			if len(out.Candidates) != 3 {
				t.Fatalf("candidates=%d, want 3", len(out.Candidates))
			}
			seen := map[string]bool{}
			for _, c := range out.Candidates {
				if c.ScenarioLabel != mode {
					t.Fatalf("scenario=%s, want %s", c.ScenarioLabel, mode)
				}
				if c.Text == "" || c.WhyThisWorks == "" || c.ToneLabel == "" {
					t.Fatalf("candidate missing display fields: %+v", c)
				}
				seen[c.Text] = true
			}
			if len(seen) != 3 {
				t.Fatalf("mock candidates should be distinct: %+v", out.Candidates)
			}
		})
	}
}

func TestParseCandidatesValidatesContract(t *testing.T) {
	req := testGenerateRequest(domain.ModeReply, "zh-CN")
	got, err := parseCandidates(`{"candidates":[
		{"text":"可以，我们慢慢聊。","scenario_label":"wrong","risk_level":"high"},
		{"text":"我想认真回应你。"},
		{"text":"谢谢你愿意说出来。","risk_level":"medium"}
	]}`, req)
	if err != nil {
		t.Fatalf("parse candidates: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("candidates=%d, want 3", len(got))
	}
	if got[0].ScenarioLabel != domain.ModeReply || got[0].RiskLevel != "low" {
		t.Fatalf("candidate not normalized: %+v", got[0])
	}
	if got[1].ToneLabel != req.Tone {
		t.Fatalf("tone=%s, want %s", got[1].ToneLabel, req.Tone)
	}

	if _, err := parseCandidates(`{"candidates":[{"text":"只有一条"}]}`, req); err == nil {
		t.Fatal("expected error for too few candidates")
	}
	if _, err := parseCandidates(`{"candidates":[{"text":""},{"text":"ok"},{"text":"ok"}]}`, req); err == nil {
		t.Fatal("expected error for empty candidate text")
	}
}

func TestAgentRegressionCases(t *testing.T) {
	cases := []domain.GenerateRequest{
		{Mode: domain.ModeRewrite, Language: "zh-CN", Draft: "你别烦我了", Tone: "gentle", Count: 3},
		{Mode: domain.ModeRewrite, Language: "en-US", Draft: "Stop bothering me", Tone: "clear", Count: 3},
		{Mode: domain.ModeReply, Language: "zh-CN", PeerMessage: "你今天怎么都不找我聊天？", Tone: "warm", Count: 3},
		{Mode: domain.ModeReply, Language: "en-US", PeerMessage: "Why did you not text me today?", Tone: "warm", Count: 3},
		{Mode: domain.ModeReply, Language: "zh-CN", PeerMessage: "你是不是不在乎我了？", Tone: "gentle", Count: 3},
		{Mode: domain.ModeOpener, Language: "zh-CN", PeerMessage: "对方朋友圈发了旅行照片", Tone: "light", Count: 3},
		{Mode: domain.ModeOpener, Language: "en-US", PeerMessage: "They posted about coffee", Tone: "light", Count: 3},
		{Mode: domain.ModeOpener, Language: "zh-CN", PeerMessage: "刚加好友，不知道说什么", Tone: "relaxed", Count: 3},
		{Mode: domain.ModeComfort, Language: "zh-CN", PeerMessage: "我最近真的很累", Tone: "gentle", Count: 3},
		{Mode: domain.ModeComfort, Language: "en-US", PeerMessage: "I feel exhausted lately", Tone: "gentle", Count: 3},
		{Mode: domain.ModeComfort, Language: "zh-CN", PeerMessage: "工作被批评了", Tone: "supportive", Count: 3},
		{Mode: domain.ModeApologize, Language: "zh-CN", Draft: "我忘了回你", Tone: "sincere", Count: 3},
		{Mode: domain.ModeApologize, Language: "en-US", Draft: "I forgot to reply", Tone: "sincere", Count: 3},
		{Mode: domain.ModeApologize, Language: "zh-CN", Draft: "刚才说重了", Tone: "sincere", Count: 3},
		{Mode: domain.ModeReject, Language: "zh-CN", PeerMessage: "今晚能不能出来？", Tone: "firm", Count: 3},
		{Mode: domain.ModeReject, Language: "en-US", PeerMessage: "Can you lend me money?", Tone: "firm", Count: 3},
		{Mode: domain.ModeReject, Language: "zh-CN", PeerMessage: "能不能帮我转发一下？", Tone: "polite", Count: 3},
		{Mode: domain.ModeReply, Language: "zh-CN", PeerMessage: "你是不是在敷衍我", Tone: "calm", Count: 3},
		{Mode: domain.ModeRewrite, Language: "zh-CN", Draft: "算了随便你", Tone: "calm", Count: 3},
		{Mode: domain.ModeReject, Language: "en-US", PeerMessage: "Can we move faster?", Tone: "gentle", Count: 3},
	}

	for i, tc := range cases {
		tc.Source = domain.SourceApp
		tc.InputPolicy = domain.InputPolicyEphemeral
		out := mockGenerate(tc, domain.ProviderConfig{Provider: "mock", Model: "mock-chat"})
		if len(out.Candidates) != 3 {
			t.Fatalf("case %d mode=%s candidates=%d, want 3", i, tc.Mode, len(out.Candidates))
		}
		prompt := BuildPrompt(tc)
		if !strings.Contains(prompt, tc.Mode) || !strings.Contains(prompt, "risk_level") {
			t.Fatalf("case %d prompt missing mode or risk contract", i)
		}
	}
}

func testGenerateRequest(mode, language string) domain.GenerateRequest {
	return domain.GenerateRequest{
		Mode:              mode,
		Source:            domain.SourceApp,
		InputPolicy:       domain.InputPolicyEphemeral,
		PeerMessage:       "你今天怎么都不找我聊天？",
		Draft:             "我不知道怎么回",
		Tone:              "gentle",
		RelationshipStage: "early_chat",
		Language:          language,
		Length:            "short",
		Count:             3,
	}
}
