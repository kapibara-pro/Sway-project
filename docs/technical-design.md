# Sway / 言和技术文档

## 1. 技术目标

Sway 技术方案采用“主 App + iOS Keyboard Extension + 后端 AI Gateway”的组合：

- 主 App 承担完整流程、账号、订阅、隐私设置和无权限兜底。
- Keyboard Extension 作为轻量输入入口，只负责用户主动提供文本后的候选生成与插入。
- 后端负责模型调用、Prompt 编排、安全策略、限流计费和指标统计。

设计重点是稳定、低延迟、隐私最小化和符合 iOS 权限边界。

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
  +-- Auth / Rate Limit / Billing
  +-- Prompt & Policy Service
  +-- Safety Filter
  +-- Model Provider
  +-- Metrics & Feedback
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
- 登录、订阅、额度展示。
- App 内输入/粘贴生成。
- 截图 OCR 导入预留。
- 收藏模板与个人风格设置。
- 隐私设置与数据保存策略。
- 问题反馈与 request_id 收集。

建议技术栈：

- Swift。
- SwiftUI 或 UIKit 均可，若项目需要快速落地可用 SwiftUI 主 App + UIKit Keyboard Extension。
- Keychain 存储敏感 token。
- App Group 存储语气偏好、模板缓存和非敏感配置。

### 4.2 Keyboard Extension

职责：

- 改写当前输入。
- 用户主动粘贴对方消息。
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
- 限流和额度。
- 模型路由。
- Prompt 模板选择。
- 内容安全检测。
- 成本、延迟和错误指标。
- 默认只记录元信息，不记录原文。

### 5.2 Prompt & Policy Service

按场景维护 Prompt、版本和安全策略：

- `rewrite`：改写用户草稿，不改变原意。
- `reply`：根据对方消息生成回复。
- `opener`：破冰开场。
- `comfort`：安慰共情。
- `apologize`：道歉修复。
- `reject`：体面拒绝。

## 6. API 契约

### 6.1 生成接口

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
  "input_policy": "ephemeral",
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
- `input_policy`：`ephemeral | store_allowed`，默认 `ephemeral`。
- `peer_message`：对方消息，可空，第一版上限 1000 字。
- `draft`：用户草稿，可空，第一版上限 500 字。
- `tone`：`gentle | humorous | flirty | sincere | concise | proactive | restrained`。
- `relationship_stage`：`stranger | early_chat | dating | couple | conflict | unknown`。
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

### 6.2 错误响应

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

客户端应按 `code` 做分流，展示本地化文案和 `fallback_suggestion`，不要直接暴露技术错误。

## 7. 数据与存储

建议表：

- `chat_assist_requests`：request_id、user_id、mode、tone、source、model、latency、usage、risk、created_at；默认不存原文。
- `chat_assist_feedback`：request_id、candidate_id、action、reason、created_at。
- `user_style_profiles`：用户语气偏好、禁用风格、常用语言。
- `saved_templates`：用户收藏话术。
- `prompt_versions`：场景 Prompt、版本、灰度配置。

数据策略：

- `input_policy=ephemeral` 时不落原文。
- `input_policy=store_allowed` 仅在用户明确授权时使用。
- 日志默认只存元信息、错误码、延迟、token、风险标签。

## 8. 安全与隐私

- HTTPS 全链路传输。
- Token 存 Keychain，避免明文落盘。
- App Group 只存非敏感配置。
- 默认不保存原始聊天内容。
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

- Full Access 未开启：提示去主 App 生成，或使用本地模板。
- 网络不可用：本地模板 + 重试。
- 模型超时：展示 `fallback_suggestion`。
- 超额度：展示额度说明和会员入口。
- 安全拒绝：展示健康表达建议。

## 11. 开发阶段

### P0：App 内生成

- 建立仓库结构。
- 实现后端生成 API。
- 主 App 内输入/粘贴生成。
- Prompt 与安全策略回归。

### P1：Keyboard Extension

- 实现键盘 UI。
- 接入 Full Access 联网。
- 接入 App Group 配置。
- 插入候选文本。
- 完成常见 App 兼容矩阵。

### P2：账号、订阅和反馈

- 用户体系。
- 会员和额度。
- 收藏模板。
- 反馈数据。

### P3：体验优化

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
