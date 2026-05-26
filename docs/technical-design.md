# Sway / 言和技术文档

## 1. 技术目标

Sway 技术方案采用“主 App + iOS Keyboard Extension + 后端 AI Gateway”的组合：

- 主 App 承担完整流程、账号、订阅、隐私设置和无权限兜底。
- Keyboard Extension 作为轻量输入入口，只负责用户主动提供文本后的候选生成与插入。
- 后端负责用户自配 LLM Provider 的代理调用、Prompt 编排、安全策略、限流和指标统计。

设计重点是稳定、低延迟、隐私最小化和符合 iOS 权限边界。

首版范围：

- iOS only。
- P0 直接包含主 App 与 iOS Keyboard Extension。
- 同时交付后端管理系统，用于用量限制、模型策略、日志、样本和错误监控。
- 六个模式进入 MVP：`rewrite | reply | opener | comfort | apologize | reject`。
- 支持用户主动复制/粘贴与截图 OCR 提供上下文。
- 账号、订阅和会员后置到 P3。
- 第一版使用游客设备身份，预留登录扩展点。
- LLM Provider 由用户在 App 内配置，首版由后端代理调用。
- 生成语言支持中文和英文。
- 历史和样本允许保存，保留 1 个月。
- 发布方式为 TestFlight 内测。

## 2. 总体架构

```text
用户
  |
  | 主动输入 / 粘贴 / 导入截图
  v
iOS 主 App --------------+
  |                      |
  | App Group / Keychain |
  v                      |
Keyboard Extension ------+
  |
  | HTTPS API
  v
Backend AI Gateway
  |
  +-- Provider Config Proxy
  +-- Auth / Rate Limit
  +-- Prompt & Policy Service
  +-- Safety Filter
  +-- Model Provider
  +-- Metrics & Feedback
  |
Admin Web
  |
  +-- Usage Limit Config
  +-- Prompt / Safety Config
  +-- Request Logs / Errors
  +-- Sample Management
```

## 3. iOS 能力边界

### 3.1 可以做

- 使用 `UIInputViewController` 实现第三方键盘。
- 使用 `textDocumentProxy.insertText` 将候选文本插入当前输入框。
- 通过 `UITextDocumentProxy` 读取当前输入框插入点附近文本，用于改写当前草稿。
- 用户主动粘贴对方消息后生成回复。
- 使用 App Group 同步非敏感配置和模板。
- 开启 Full Access 后由键盘扩展发起网络请求。

### 3.2 不能做

- 不能读取微信、QQ、Telegram 等第三方 App 的完整聊天记录。
- 不能读取屏幕上的历史消息列表。
- 不能自动监听新消息。
- 不能自动点击发送按钮。
- 不能绕过系统沙盒读取第三方 App 数据。
- 不能在安全输入框中继续使用第三方键盘。

### 3.3 Full Access

Keyboard Extension 默认没有网络访问和共享容器写入能力。若需要键盘内直接调用 AI 服务，需要：

- 在扩展 `Info.plist` 中配置 `RequestsOpenAccess = true`。
- 用户在系统设置中开启“允许完全访问”。
- 在产品和隐私文案中说明开启权限的用途。

Full Access 只代表允许键盘联网和访问共享容器，不代表可以读取宿主 App 的聊天记录。

### 3.4 剪贴板

iOS 16 之后，跨 App 读取剪贴板可能触发系统粘贴授权提示。实现上应避免后台轮询剪贴板，采用用户点击“粘贴/导入”后的显式读取。

## 4. iOS 模块设计

### 4.1 主 App

职责：

- Onboarding 与输入法安装引导。
- Full Access 权限解释。
- Provider 配置、连通性测试和配置清除。
- App 内输入/粘贴生成。
- 截图 OCR 导入。
- 中英文生成。
- 游客设备身份，预留登录入口。
- 收藏模板与个人风格设置。
- 隐私设置、历史/样本保存说明和清除入口。
- 问题反馈与 request_id 收集。

建议技术栈：

- Swift。
- SwiftUI 或 UIKit 均可，若项目需要快速落地可用 SwiftUI 主 App + UIKit Keyboard Extension。
- Keychain 存储设备标识、会话凭证等敏感本地信息。
- App Group 存储语气偏好、模板缓存和非敏感配置。

### 4.2 Keyboard Extension

职责：

- 改写当前输入。
- 用户主动粘贴对方消息。
- 处理截图 OCR/主 App 导入的上下文兜底。
- 选择语气和场景。
- 展示 3 条候选。
- 插入候选文本。
- 错误和降级提示。

约束：

- UI 必须轻量，避免复杂导航。
- 请求超时控制在 4-5 秒。
- 不保存敏感长历史。
- 不直接展示服务端技术错误。

## 5. 后端模块设计

### 5.1 Backend AI Gateway

职责：

- API 鉴权。
- 用户设备/安装 ID 识别。
- Provider 配置加密存储。
- 限流。
- 用量限制配置与检查。
- 模型代理调用。
- Prompt 模板选择。
- 内容安全检测。
- 成本、延迟和错误指标。
- 历史和样本 1 个月 TTL 存储。
- 后端管理 API。

### 5.2 Admin Web

首版后台管理前端用于内测运营和质量观察，不面向普通用户。

职责：

- 配置用量限制：免费次数、频控、按设备/模型/模式统计。
- 管理模型和 Prompt 策略。
- 查看请求日志、usage、错误码、延迟和 token。
- 查看和清理 1 个月内样本，敏感字段需要脱敏展示。
- 查看安全策略命中和降级原因。

### 5.3 Prompt & Policy Service

按场景维护 Prompt、版本和安全策略：

- `rewrite`：改写用户草稿，不改变原意。
- `reply`：根据对方消息生成回复。
- `opener`：破冰开场。
- `comfort`：安慰共情。
- `apologize`：道歉修复。
- `reject`：体面拒绝。

## 6. API 契约

### 6.1 Provider 配置接口

客户端通过主 App 配置用户自有 LLM Provider。Keyboard Extension 只读取配置状态，未配置时提示回主 App。

```http
GET /api/v1/providers/status
```

返回是否已配置、provider、model、base_url 掩码展示、最后校验时间，不返回 API Key。

```http
PUT /api/v1/providers/config
Content-Type: application/json
```

请求示例：

```json
{
  "provider": "openai",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-4.1-mini",
  "api_key": "sk-..."
}
```

服务端必须加密存储 API Key，日志不得打印 API Key，状态接口只能返回掩码。

```http
POST /api/v1/providers/test
```

使用一条短 prompt 做连通性测试，返回统一错误码。

```http
DELETE /api/v1/providers/config
```

清除用户 Provider 配置。

### 6.2 生成接口

```http
POST /api/v1/chat-assist/generate
Content-Type: application/json
Authorization: Bearer <token>
```

请求示例：

```json
{
  "mode": "reply",
  "source": "keyboard",
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
    "extension_version": "0.1.0",
    "locale": "zh-CN"
  }
}
```

字段说明：

- `mode`：`rewrite | reply | opener | comfort | apologize | reject`。
- `source`：`keyboard | app`。
- `input_policy`：`ephemeral | store_allowed`，首版允许 `store_allowed`，历史与样本保留 1 个月。
- `peer_message`：对方消息，可空，第一版上限 1000 字。
- `draft`：用户草稿，可空，第一版上限 500 字。
- `tone`：`gentle | humorous | flirty | sincere | concise | proactive | restrained`。
- `relationship_stage`：`stranger | early_chat | dating | couple | conflict | unknown`。
- `language`：`zh-CN | en-US`，首版支持中文和英文。
- `length`：`short | medium`。
- `count`：默认 3，第一版固定返回 3 条。

总上下文第一版限制为 1500 字以内。

响应示例：

```json
{
  "request_id": "req_20260526_000001",
  "candidates": [
    {
      "id": "c1",
      "text": "今天有点忙，但我其实一直惦记着你。现在来补上。",
      "tone_label": "温柔",
      "scenario_label": "安抚",
      "risk_level": "low"
    }
  ],
  "safety": {
    "blocked": false,
    "reason": null
  },
  "usage": {
    "model": "fast-chat-model",
    "latency_ms": 1280,
    "tokens": 420
  }
}
```

候选文案建议每条控制在 80 字以内。

### 6.3 错误响应

```json
{
  "code": "TEXT_TOO_LONG",
  "message": "Input text is too long.",
  "fallback_suggestion": "内容有点长，可以删短一点再试。",
  "request_id": "req_20260526_000002",
  "details": {
    "field": "peer_message",
    "limit": 1000
  }
}
```

第一版错误码：

- `INVALID_INPUT`：空文本、非法 mode/tone、缺少必要字段。
- `TEXT_TOO_LONG`：输入超长。
- `FULL_ACCESS_REQUIRED`：键盘侧无联网权限，需要跳主 App。
- `SAFETY_BLOCKED`：内容安全拒绝或降级。
- `MODEL_TIMEOUT`：模型超时。
- `RATE_LIMITED`：额度或频控限制。
- `SERVICE_UNAVAILABLE`：后端临时不可用。
- `PROVIDER_NOT_CONFIGURED`：用户还没有配置 LLM Provider。
- `PROVIDER_AUTH_FAILED`：Provider API Key 无效或权限不足。
- `PROVIDER_TEST_FAILED`：Provider 连通性测试失败。

客户端应按 `code` 做分流，展示本地化文案和 `fallback_suggestion`，不要直接暴露技术错误。

## 7. 数据与存储

建议表：

- `chat_assist_requests`：request_id、user_id/device_id、mode、tone、source、model、latency、usage、risk、created_at。
- `chat_assist_feedback`：request_id、candidate_id、action、reason、created_at。
- `user_style_profiles`：用户语气偏好、禁用风格、常用语言。
- `saved_templates`：用户收藏话术。
- `prompt_versions`：场景 Prompt、版本、灰度配置。
- `provider_configs`：用户 Provider、base_url、model、API Key 密文、最后校验时间。
- `chat_histories`：用户生成历史，按 1 个月 TTL 清理。
- `quality_samples`：允许保存的样本，按 1 个月 TTL 清理。
- `usage_limits`：用量限制、频控、适用范围和生效状态。
- `admin_audit_logs`：后台操作记录。

数据策略：

- 历史与样本保留 1 个月。
- App 内提供清除入口。
- `input_policy=ephemeral` 时不落原文。
- `input_policy=store_allowed` 时按 1 个月 TTL 保存。
- 日志默认只存元信息、错误码、延迟、token、风险标签。

## 8. 安全与隐私

- HTTPS 全链路传输。
- iOS 本地敏感信息存 Keychain，避免明文落盘。
- Provider API Key 在服务端加密存储，接口响应不得明文返回。
- 服务端日志不得打印 API Key。
- App Group 只存非敏感配置。
- 历史与样本保存 1 个月，提供清除入口。
- 不后台读取剪贴板。
- 不自动读取通讯录。
- 不承诺读取第三方聊天 App 完整上下文。
- App Store 隐私政策和 `PrivacyInfo.xcprivacy` 需要覆盖数据收集、用途、第三方 SDK 和保留策略。

## 9. 性能目标

- 键盘内生成 p95 控制在 2.5-3 秒。
- Keyboard Extension API 超时 4-5 秒。
- 一次模型调用返回 3 条候选，避免多次调用放大延迟和成本。
- 失败时展示本地模板、重试或跳主 App。

## 10. 降级策略

- Provider 未配置：提示回主 App 配置模型。
- Full Access 未开启：提示去主 App 生成，或使用本地模板。
- 网络不可用：本地模板 + 重试。
- OCR 失败：允许用户手动编辑识别文本或改用粘贴。
- 模型超时：展示 `fallback_suggestion`。
- 频控限制：提示稍后重试。
- 安全拒绝：展示健康表达建议。

## 11. 开发阶段

### P0：iOS 内测 MVP

- 建立仓库结构。
- iOS 主 App。
- iOS Keyboard Extension。
- 后台管理前端。
- Provider 配置页与连通性测试。
- 实现后端生成 API。
- 实现后端管理 API。
- 主 App 内输入/粘贴生成。
- 截图 OCR 导入。
- 键盘内改写、回复和插入候选。
- Prompt 与安全策略回归。
- 六个模式与中英文生成。
- 游客设备身份与登录预留。
- 用量限制后台配置。
- 历史与样本 1 个月 TTL。

### P1：输入法体验完善

- Full Access 引导和异常兜底。
- 完成常见 App 兼容矩阵。
- 主流 iPhone 和 iPad 适配。
- TestFlight 内测反馈修复。

### P2：反馈与质量

- 收藏模板。
- 反馈数据。
- 质量、延迟、成本看板。

### P3：账号与商业化

- 用户体系。
- 会员和额度。
- IAP。
- 商业化实验。

### P4：体验优化

- 个性化语气。
- 多语言。
- 本地模板增强。
- Prompt A/B。

## 12. 兼容性测试矩阵

第一版至少覆盖：

- 微信。
- QQ。
- 短信。
- Telegram。
- WhatsApp。
- 小红书私信。
- 抖音私信。
- 备忘录。
- Safari 输入框。

测试项：

- 键盘是否可唤起。
- 当前草稿是否可读取。
- 候选是否可插入。
- Full Access 开启前后表现。
- 剪贴板提示表现。
- 安全输入框降级表现。
- 生成失败和超时文案。
