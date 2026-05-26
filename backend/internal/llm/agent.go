package llm

import (
	"fmt"
	"strings"

	"github.com/kapibara-pro/sway-project/backend/internal/domain"
)

type modeSpec struct {
	NameZH       string
	NameEN       string
	GoalZH       string
	GoalEN       string
	Constraints  []string
	StrategiesZH []string
	StrategiesEN []string
}

var modeSpecs = map[string]modeSpec{
	domain.ModeRewrite: {
		NameZH: "改写草稿", NameEN: "rewrite draft",
		GoalZH: "在不改变用户原意、不新增事实的前提下，把草稿改得更自然、更有分寸。",
		GoalEN: "Make the user's draft more natural and considerate without changing intent or adding facts.",
		Constraints: []string{
			"保留用户原意，不替用户承诺没有说过的事。",
			"不要把短消息改成作文；优先保持聊天口吻。",
		},
		StrategiesZH: []string{"更温柔", "更松弛", "更清楚"},
		StrategiesEN: []string{"warmer", "lighter", "clearer"},
	},
	domain.ModeReply: {
		NameZH: "根据对方消息回复", NameEN: "reply to a message",
		GoalZH: "理解对方消息的情绪和关系语境，给出可以直接复制的自然回复。",
		GoalEN: "Read the other person's tone and relationship context, then draft natural replies the user can copy.",
		Constraints: []string{
			"不要替用户操控、施压、诱导或制造愧疚感。",
			"遇到暧昧场景也要保持尊重边界，不越界升级关系。",
		},
		StrategiesZH: []string{"安全回应", "高情商回应", "轻推进回应"},
		StrategiesEN: []string{"safe reply", "emotionally aware reply", "gentle next-step reply"},
	},
	domain.ModeOpener: {
		NameZH: "破冰开场", NameEN: "icebreaker",
		GoalZH: "基于用户提供的线索生成轻松、不冒犯、方便对方接话的开场。",
		GoalEN: "Create a light, respectful opener based on the user's context that is easy to answer.",
		Constraints: []string{
			"不要油腻搭讪，不要强行亲密。",
			"没有足够上下文时，给通用但不尴尬的开场。",
		},
		StrategiesZH: []string{"轻松开场", "好奇提问", "低压力推进"},
		StrategiesEN: []string{"light opener", "curious question", "low-pressure nudge"},
	},
	domain.ModeComfort: {
		NameZH: "安慰共情", NameEN: "comfort",
		GoalZH: "先承接情绪，再给轻量支持，避免说教和无效鸡汤。",
		GoalEN: "Acknowledge emotion first, then offer gentle support without lecturing or empty clichés.",
		Constraints: []string{
			"不要诊断、评判或要求对方立刻好起来。",
			"不要用模板化的“我理解你的感受”开头。",
		},
		StrategiesZH: []string{"共情承接", "陪伴支持", "轻建议"},
		StrategiesEN: []string{"empathic acknowledgement", "steady support", "small practical support"},
	},
	domain.ModeApologize: {
		NameZH: "道歉修复", NameEN: "apologize and repair",
		GoalZH: "承担该承担的责任，解释但不甩锅，并给出可执行的修复动作。",
		GoalEN: "Take appropriate responsibility, explain without deflecting, and offer a concrete repair.",
		Constraints: []string{
			"不要过度卑微，也不要把责任推给对方。",
			"没有事实依据时，不编造原因。",
		},
		StrategiesZH: []string{"真诚道歉", "解释修复", "关系缓和"},
		StrategiesEN: []string{"sincere apology", "repair with context", "relationship repair"},
	},
	domain.ModeReject: {
		NameZH: "体面拒绝", NameEN: "respectful refusal",
		GoalZH: "清楚表达边界，同时尽量保留体面和尊重。",
		GoalEN: "State the boundary clearly while preserving dignity and respect where possible.",
		Constraints: []string{
			"不要给模糊希望，不要为了讨好而答应。",
			"拒绝应简洁明确，不攻击对方。",
		},
		StrategiesZH: []string{"温和拒绝", "清晰边界", "给替代选项"},
		StrategiesEN: []string{"gentle no", "clear boundary", "alternative option"},
	},
}

// BuildSystemPrompt returns the stable system contract for the LLM provider.
func BuildSystemPrompt(req domain.GenerateRequest) string {
	lang := "中文"
	if req.Language == "en-US" {
		lang = "English"
	}
	return fmt.Sprintf(`你是 Sway/言和的高情商聊天表达 Agent，只做表达增强，不替用户聊天。

核心边界：
- 只生成候选文本，用户必须自己确认和发送。
- 不操控、不PUA、不骚扰、不制造愧疚感，不替用户越界承诺。
- 高风险内容要改成健康、尊重边界的表达；不能提供跟踪、威胁、性化未成年人、自伤鼓励等内容。
- 输出必须是严格 JSON，不要输出 Markdown、解释段落或多余文本。
- 用户可见内容必须使用 %s。

JSON schema：
{"candidates":[{"text":"候选文案","tone_label":"语气标签","scenario_label":"%s","risk_level":"low|medium","why_this_works":"为什么这条适合当前语境"}]}`, lang, req.Mode)
}

// BuildPrompt creates a mode-specific user prompt for the generation request.
func BuildPrompt(req domain.GenerateRequest) string {
	spec := specForMode(req.Mode)
	goal := spec.GoalZH
	strategies := spec.StrategiesZH
	if req.Language == "en-US" {
		goal = spec.GoalEN
		strategies = spec.StrategiesEN
	}
	return fmt.Sprintf(`任务模式: %s (%s)
目标: %s
语气: %s
关系阶段: %s
长度: %s
语言: %s
候选数量: %d

对方消息:
%s

我的草稿:
%s

模式约束:
%s

请生成 %d 条不同策略的候选，策略分别偏向:
%s

输出要求:
- 只能返回 JSON: {"candidates":[...]}。
- 每条 candidate 必须包含 text、tone_label、scenario_label、risk_level、why_this_works。
- scenario_label 固定为 "%s"。
- text 尽量不超过 80 字，像真人聊天，不要 AI 腔。
- 如果用户输入里有操控、施压、骚扰、过度承诺等风险，把候选改成更健康的表达，risk_level 用 "medium"。
- 不要生成自动发送、读取聊天记录、冒充用户已做决定的内容。`,
		req.Mode,
		displayName(spec, req.Language),
		goal,
		emptyDefault(req.Tone, "gentle"),
		emptyDefault(req.RelationshipStage, "unknown"),
		emptyDefault(req.Length, "short"),
		emptyDefault(req.Language, "zh-CN"),
		countFor(req),
		emptyDefault(req.PeerMessage, "(empty)"),
		emptyDefault(req.Draft, "(empty)"),
		joinBullets(spec.Constraints),
		countFor(req),
		joinBullets(strategies),
		req.Mode,
	)
}

func mockGenerate(req domain.GenerateRequest, cfg domain.ProviderConfig) GenerateOutput {
	count := countFor(req)
	candidates := mockCandidates(req)
	if len(candidates) > count {
		candidates = candidates[:count]
	}
	for i := range candidates {
		candidates[i].ID = fmt.Sprintf("c%d", i+1)
		if candidates[i].ToneLabel == "" {
			candidates[i].ToneLabel = emptyDefault(req.Tone, "gentle")
		}
		if candidates[i].ScenarioLabel == "" {
			candidates[i].ScenarioLabel = req.Mode
		}
		if candidates[i].RiskLevel == "" {
			candidates[i].RiskLevel = "low"
		}
	}
	return GenerateOutput{
		Candidates: candidates,
		Usage: domain.Usage{
			Provider:         cfg.Provider,
			Model:            cfg.Model,
			PromptTokens:     220,
			CompletionTokens: 120,
			TotalTokens:      340,
		},
	}
}

func mockCandidates(req domain.GenerateRequest) []domain.Candidate {
	if req.Language == "en-US" {
		return mockCandidatesEN(req.Mode)
	}
	return mockCandidatesZH(req.Mode)
}

func mockCandidatesZH(mode string) []domain.Candidate {
	why := map[string]string{
		domain.ModeRewrite:   "保留原意，同时把表达变得更自然、有分寸。",
		domain.ModeReply:     "先接住对方情绪，再给出轻松可接的话口。",
		domain.ModeOpener:    "降低开场压力，让对方容易回复。",
		domain.ModeComfort:   "先承接情绪，再给陪伴感，不急着说教。",
		domain.ModeApologize: "承担责任并给出修复动作，避免甩锅。",
		domain.ModeReject:    "边界清楚，同时保留体面和尊重。",
	}
	texts := map[string][]string{
		domain.ModeRewrite: {
			"我想认真说一下：我在意你的感受，也希望我们能舒服地聊下去。",
			"刚才那句话我想换个说法，我不是想让你有压力，只是想把心意说清楚。",
			"我有点在意这件事，但不想把气氛弄重，我们慢慢说就好。",
		},
		domain.ModeReply: {
			"没有不想找你，只是今天有点忙。看到你这么说，我其实还挺开心的。",
			"你这么问我有点心软。那我现在认真补一句：今天也有想到你。",
			"是我节奏慢了点，不是冷淡你。我们现在聊也刚刚好。",
		},
		domain.ModeOpener: {
			"刚看到你发的内容，感觉挺有意思的。这个是你最近一直在关注的吗？",
			"冒昧打个招呼，你这个点我还挺好奇的，想听你多讲两句。",
			"我不太会开场，但这个话题确实吸引到我了。",
		},
		domain.ModeComfort: {
			"听起来你这阵子真的撑得挺累的。先别急着把自己调整好，我在这儿听你说。",
			"这件事换谁都会不好受。你不用马上变得没事，我们可以慢慢来。",
			"我不急着给建议，只想先陪你把这口气缓下来。",
		},
		domain.ModeApologize: {
			"这件事是我处理得不够好，让你不舒服了。对不起，我会把后面的动作补上。",
			"我不想用忙来当借口。确实是我忽略了你的感受，我会认真改。",
			"对不起，刚才那样说不合适。谢谢你愿意提醒我，我会注意分寸。",
		},
		domain.ModeReject: {
			"谢谢你这么说，但这件事我可能不能答应。还是想直接一点，不让你误会。",
			"我理解你的想法，不过我现在不太适合继续往这个方向走。",
			"这次我就先不参与了。希望你别介意，也祝你顺利。",
		},
	}
	spec := specForMode(mode)
	out := make([]domain.Candidate, 0, 3)
	for i, text := range texts[mode] {
		strategy := spec.StrategiesZH[i%len(spec.StrategiesZH)]
		out = append(out, domain.Candidate{Text: text, ToneLabel: strategy, ScenarioLabel: mode, RiskLevel: "low", WhyThisWorks: why[mode]})
	}
	if len(out) == 0 {
		return mockCandidatesZH(domain.ModeReply)
	}
	return out
}

func mockCandidatesEN(mode string) []domain.Candidate {
	why := map[string]string{
		domain.ModeRewrite:   "Keeps the intent while making the wording warmer and more natural.",
		domain.ModeReply:     "Acknowledges the emotion and leaves an easy opening to continue.",
		domain.ModeOpener:    "Starts lightly without making the other person feel pressured.",
		domain.ModeComfort:   "Acknowledges the feeling before offering calm support.",
		domain.ModeApologize: "Takes responsibility and offers a repair without deflecting.",
		domain.ModeReject:    "Sets a clear boundary while staying respectful.",
	}
	texts := map[string][]string{
		domain.ModeRewrite: {
			"I want to say this more clearly: I care about how you feel, and I want us to talk comfortably.",
			"Let me put that a little better. I am not trying to pressure you; I just want to be honest.",
			"This matters to me, but I do not want to make it heavy. We can talk it through slowly.",
		},
		domain.ModeReply: {
			"I did not mean to disappear. Today got a bit full, but seeing your message honestly made me smile.",
			"That made me soften a little. So here is a proper reply: I did think about you today.",
			"My pace was slow, not cold. We can still talk now if you are here.",
		},
		domain.ModeOpener: {
			"I saw what you posted and got curious. Is this something you have been into lately?",
			"Small hello from me. That point caught my attention, and I would like to hear more.",
			"I am not great at openers, but this topic genuinely made me want to ask.",
		},
		domain.ModeComfort: {
			"That sounds like a lot to carry. You do not have to fix it immediately; I am here to listen.",
			"Anyone would feel shaken by that. You do not need to be okay right away.",
			"I will not rush into advice. I just want to sit with you in this for a moment.",
		},
		domain.ModeApologize: {
			"I handled that poorly, and I can see why it hurt. I am sorry, and I will make it right.",
			"I do not want to hide behind being busy. I missed how that landed, and I will be more careful.",
			"Sorry, that was not a fair way to say it. Thank you for telling me; I will adjust.",
		},
		domain.ModeReject: {
			"Thank you for asking, but I cannot say yes to this. I would rather be clear than leave you guessing.",
			"I understand where you are coming from, but this is not something I can move forward with.",
			"I will pass this time. I hope you understand, and I still wish you well.",
		},
	}
	spec := specForMode(mode)
	out := make([]domain.Candidate, 0, 3)
	for i, text := range texts[mode] {
		strategy := spec.StrategiesEN[i%len(spec.StrategiesEN)]
		out = append(out, domain.Candidate{Text: text, ToneLabel: strategy, ScenarioLabel: mode, RiskLevel: "low", WhyThisWorks: why[mode]})
	}
	if len(out) == 0 {
		return mockCandidatesEN(domain.ModeReply)
	}
	return out
}

func specForMode(mode string) modeSpec {
	if spec, ok := modeSpecs[mode]; ok {
		return spec
	}
	return modeSpecs[domain.ModeReply]
}

func countFor(req domain.GenerateRequest) int {
	if req.Count <= 0 || req.Count > 3 {
		return 3
	}
	return req.Count
}

func displayName(spec modeSpec, language string) string {
	if language == "en-US" {
		return spec.NameEN
	}
	return spec.NameZH
}

func joinBullets(items []string) string {
	if len(items) == 0 {
		return "- 无"
	}
	var b strings.Builder
	for _, item := range items {
		b.WriteString("- ")
		b.WriteString(item)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func emptyDefault(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
