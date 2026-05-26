# Sway Backend API v1

M1 后端目标是先跑通主 App 与 Keyboard Extension 共用的生成闭环：游客设备身份、Provider 配置、Provider 连通性测试、生成 API、统一错误码、request/usage 记录。

M2 在保持接口兼容的基础上，将生成链路升级为六模式高情商 LLM Agent：客户端仍只传 `mode/language/tone/input_policy/peer_message/draft` 等结构化字段，后端负责按模式拼装 prompt、调用用户配置的 OpenAI-compatible Provider，并校验返回的 3 条候选。

## 1. 本地启动

```bash
cd backend
go run ./cmd/server
```

默认监听 `:8080`，可通过环境变量覆盖：

```bash
SWAY_BACKEND_ADDR=:8080 go run ./cmd/server
```

真实 LLM 生成默认超时为 30 秒，可通过环境变量覆盖：

```bash
SWAY_LLM_TIMEOUT_SECONDS=30 go run ./cmd/server
```

说明：后端会请求 `${base_url}/chat/completions`。例如 `base_url=https://dashscope-intl.aliyuncs.com/compatible-mode/v1` 时，实际请求地址为 `https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions`。如果模型生成或首包超过 `SWAY_LLM_TIMEOUT_SECONDS`，接口会返回 `MODEL_TIMEOUT`。

健康检查：

```http
GET /healthz
```

## 2. 游客设备身份

M1 不做账号系统。客户端每次请求都带设备游客 ID：

```http
X-Sway-Device-ID: <install_id_or_device_id>
```

后续登录体系上线后，可以把该 ID 绑定到正式账号。

## 3. Provider 配置

### 3.1 查看配置状态

```http
GET /api/v1/providers/status
X-Sway-Device-ID: device-demo
```

未配置：

```json
{
  "configured": false,
  "last_test_ok": false
}
```

已配置：

```json
{
  "configured": true,
  "provider": "mock",
  "model": "mock-chat",
  "key_mask": "****",
  "updated_at": "2026-05-26T07:30:00Z",
  "last_test_ok": true
}
```

### 3.2 保存配置

```http
PUT /api/v1/providers/config
Content-Type: application/json
X-Sway-Device-ID: device-demo
```

真实 OpenAI-compatible Provider：

```json
{
  "provider": "openai",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-4.1-mini",
  "api_key": "sk-..."
}
```

M1 本地联调可用 mock Provider，不会调用外部模型：

```json
{
  "provider": "mock",
  "base_url": "",
  "model": "mock-chat",
  "api_key": ""
}
```

注意：状态接口只返回 `key_mask`，不会返回 API Key 明文。

### 3.3 测试配置

```http
POST /api/v1/providers/test
X-Sway-Device-ID: device-demo
```

成功：

```json
{
  "ok": true,
  "status": {
    "configured": true,
    "provider": "mock",
    "model": "mock-chat",
    "key_mask": "****",
    "last_test_ok": true
  }
}
```

### 3.4 清除配置

```http
DELETE /api/v1/providers/config
X-Sway-Device-ID: device-demo
```

## 4. 生成接口

```http
POST /api/v1/chat-assist/generate
Content-Type: application/json
X-Sway-Device-ID: device-demo
```

请求：

```json
{
  "mode": "reply",
  "source": "app",
  "input_policy": "store_allowed",
  "peer_message": "你今天怎么都不找我聊天？",
  "draft": "",
  "tone": "gentle",
  "relationship_stage": "early_chat",
  "language": "zh-CN",
  "length": "short",
  "count": 3,
  "client_context": {
    "platform": "ios",
    "app_version": "0.1.0",
    "locale": "zh-CN"
  }
}
```

约束：

- `mode`: `rewrite | reply | opener | comfort | apologize | reject`
- `source`: `keyboard | app`
- `input_policy`: `ephemeral | store_allowed`
- `peer_message`: 最多 1000 字
- `draft`: 最多 500 字
- 总上下文最多 1500 字
- `language`: `zh-CN | en-US`
- `count`: 1-3，M1 固定建议传 3

返回：

```json
{
  "request_id": "req_xxx",
  "candidates": [
    {
      "id": "c1",
      "text": "我明白你的意思，也想认真回应你。我们可以慢慢说，不急着把话讲满。",
      "tone_label": "gentle",
      "scenario_label": "reply",
      "risk_level": "low",
      "why_this_works": "保留原意，同时让表达更自然、有分寸。"
    }
  ],
  "safety": {
    "blocked": false
  },
  "usage": {
    "provider": "mock",
    "model": "mock-chat",
    "prompt_tokens": 120,
    "completion_tokens": 80,
    "total_tokens": 200,
    "latency_ms": 1
  }
}
```

### 4.1 六模式 Agent 规则

后端会按 `mode` 使用独立 Agent 策略：

| mode | 目标 | 关键约束 |
| --- | --- | --- |
| `rewrite` | 改写用户草稿 | 保留原意，不新增事实，不替用户承诺 |
| `reply` | 根据对方消息回复 | 接住情绪，避免操控、施压或制造愧疚 |
| `opener` | 破冰开场 | 轻松、不油腻、不强行亲密 |
| `comfort` | 安慰共情 | 先承接情绪，不说教、不诊断 |
| `apologize` | 道歉修复 | 承担责任，解释但不甩锅，给修复动作 |
| `reject` | 体面拒绝 | 边界明确，不给模糊希望，不攻击对方 |

Provider 必须返回严格 JSON：

```json
{
  "candidates": [
    {
      "text": "候选文案",
      "tone_label": "更温柔",
      "scenario_label": "reply",
      "risk_level": "low",
      "why_this_works": "先接住对方情绪，再给出轻松可接的话口。"
    }
  ]
}
```

M2 mock provider 已改成六模式专属候选，便于 iOS 在无真实 LLM 时联调。真实 Provider 使用同一响应契约。

## 5. Request / Usage 记录

M1 提供最小查询接口，便于 iOS 联调和后台页面接入：

```http
GET /api/v1/chat-assist/requests?device_id=device-demo
```

- `input_policy=ephemeral`：只保存长度、usage、错误码、延迟等元信息，不保存原文。
- `input_policy=store_allowed`：保存原文、候选和样本，`expires_at` 为 1 个月后。

## 6. 统一错误码

| code | 含义 | 客户端处理 |
| --- | --- | --- |
| `INVALID_INPUT` | 空文本、非法 mode/tone/language、缺设备 ID、未配置 Provider | 本地提示用户修正 |
| `TEXT_TOO_LONG` | 输入超长 | 提示缩短内容 |
| `FULL_ACCESS_REQUIRED` | 键盘侧需要完全访问 | 引导回主 App 或系统设置 |
| `SAFETY_BLOCKED` | 内容安全拒绝 | 展示健康表达建议 |
| `MODEL_TIMEOUT` | 模型超时 | 展示重试或本地兜底 |
| `RATE_LIMITED` | 额度/频控限制 | 展示额度或会员入口，M1 暂未启用 |
| `SERVICE_UNAVAILABLE` | Provider 或后端临时不可用 | 稍后重试 |

错误响应统一包含：

```json
{
  "code": "INVALID_INPUT",
  "message": "provider not configured",
  "fallback_suggestion": "请先在 App 中配置模型",
  "request_id": "req_xxx",
  "details": {
    "field": "mode"
  }
}
```

当前安全策略：

- 威胁、跟踪、骚扰、未成年性化、自伤鼓励等高风险内容直接返回 `SAFETY_BLOCKED`。
- 操控、施压、PUA、过度承诺等边界风险由 prompt 要求改写为健康表达，候选使用 `risk_level=medium`。
