# Sway / 言和开发排期

## 1. 最新排期输入

- 用户端只做 iOS。
- 第一版同时做 iOS 主 App、Keyboard Extension 和后端管理系统。
- 六个模式进入 MVP：改写草稿、根据对方消息回复、破冰、安慰、道歉、拒绝。
- 第一版游客模式，预留登录扩展点。
- 用量限制由后台管理系统配置。
- LLM Provider 支持 App 内配置，后端代理调用。
- 生成语言支持中文和英文。
- 接受 iOS 限制：无法自动读取微信完整聊天，只支持复制、粘贴、截图 OCR 等用户主动提供的上下文。
- 品牌定为「言和 / Sway」，需要 Logo、主色、字体/图标和 App Icon，风格为温柔高级、年轻可爱。
- 发布方式为 TestFlight/自用。

## 2. 团队分工

- @乔布斯：产品设计、信息架构、低保真流程、品牌方向。
- @戴铭：iOS 主 App、Keyboard Extension、后台管理前端、前端体验、TestFlight。
- @求伯君：后端 API、Provider 代理、Prompt/安全策略、历史与样本存储、管理后台 API。

## 3. 建议里程碑

### M0：设计与接口冻结（2 天）

目标：避免 iOS、后台前端和后端并行开发时接口反复变动。

任务：

- 产品侧输出主 App、Keyboard Extension、后台管理低保真。
- 品牌侧确认 Logo 方向、主色、基础字体/图标风格、App Icon 方向。
- iOS 与后端确认接口契约：Provider 配置、生成接口、历史/样本、用量限制、错误码。
- 确认 TestFlight 基础信息：Bundle ID、Apple Developer 账号、证书、测试设备清单。

交付物：

- 低保真设计稿。
- API 契约 v1。
- 视觉方向 v1。
- 开发任务清单。

### M1：基础工程与后端主链路（3-4 天）

目标：跑通游客设备身份、Provider 配置和生成 API。

iOS / 前端：

- 创建 iOS 主 App target。
- 创建 Keyboard Extension target。
- 配置 App Group、Keychain、本地配置模块。
- 主 App 基础导航：生成、历史、设置。
- Provider 配置页：状态、编辑、测试、清除。
- 后台管理前端工程初始化。

后端：

- 游客设备 ID。
- Provider 配置 API。
- Provider 连通性测试。
- 生成 API v1。
- 统一错误码。
- 基础 Prompt 与安全策略。

验收：

- 主 App 能保存 Provider 配置并测试连通。
- App 内能调用生成 API 返回 3 条候选。
- Keyboard Extension 工程可运行。

### M2：主 App MVP（3-4 天）

目标：主 App 完成可用的聊天助手闭环。

iOS：

- 六个模式入口。
- 中文/英文生成选择。
- 输入/粘贴上下文。
- 截图 OCR 导入与文本编辑。
- 候选展示、复制、收藏。
- 历史记录与清除入口。
- 隐私说明。

后端：

- 历史与样本保存 1 个月 TTL。
- 反馈/收藏数据接口。
- 中英文 Prompt 回归。

验收：

- 六个模式都能在主 App 内生成 3 条候选。
- OCR 导入后可编辑文本并生成。
- 历史和样本可保存、可清理。

### M3：Keyboard Extension MVP（4-5 天）

目标：在真实聊天 App 中完成输入法生成与插入。

iOS：

- Keyboard Extension UI：模式、语气、生成按钮、3 条候选、错误态。
- Full Access 检测与引导。
- 读取当前输入框草稿。
- 用户主动粘贴对方消息。
- 调用生成 API。
- 插入候选文本。
- Provider 未配置、无 Full Access、超时、安全拒绝等兜底态。

后端：

- 键盘请求来源统计。
- 短超时和降级响应。
- usage、错误码、延迟指标。

验收：

- 微信、短信、Telegram、WhatsApp、备忘录中可唤起键盘。
- 候选可稳定插入，不自动发送。
- 无 Full Access 时可引导回主 App。

### M4：后台管理 MVP（3-4 天）

目标：内测期间能配置、观察和排查。

后台前端：

- 用量限制配置页。
- Provider/模型策略页。
- Prompt/安全策略页。
- 请求日志与 usage 页面。
- 错误监控页面。
- 样本查看与清理页面。

后端：

- 管理后台 API。
- 用量限制规则生效。
- 管理操作审计。
- 样本脱敏展示。

验收：

- 后台能调整用量限制并影响生成 API。
- 能按 request_id 查看错误、耗时、token 和风险标签。
- 样本可查看、脱敏和清理。

### M5：兼容测试与 TestFlight（3-4 天）

目标：交付可内测版本。

任务：

- 主流 iPhone 和 iPad 适配。
- 微信、QQ、短信、Telegram、WhatsApp、小红书私信、抖音私信、备忘录、Safari 输入框兼容测试。
- Full Access 开启/关闭测试。
- OCR 准确率和编辑流程测试。
- 中英文生成质量回归。
- 崩溃日志与基础诊断。
- TestFlight 打包与发布。

验收：

- TestFlight 包可安装。
- 关键链路无阻塞问题。
- 已知问题清单可追踪。

## 4. iOS / 前端任务清单

### iOS 主 App

- 工程初始化与 target 配置。
- App Group 与 Keychain 封装。
- 游客设备 ID。
- Provider 配置页。
- 生成页：模式、语气、语言、上下文输入。
- 截图 OCR 导入。
- 候选展示、复制、收藏。
- 历史页与清除入口。
- 设置页：输入法引导、Full Access 说明、隐私说明。
- TestFlight 配置。

### Keyboard Extension

- 键盘 UI 框架。
- Full Access 状态检测与引导。
- 当前草稿读取。
- 粘贴对方消息。
- 生成 API 调用。
- 候选插入。
- 错误态和兜底态。
- 常见 App 兼容测试。

### 后台管理前端

- 登录占位或内测访问控制。
- Dashboard。
- 用量限制配置。
- 模型/Provider 策略。
- Prompt/安全策略。
- 请求日志、usage、错误监控。
- 样本管理与清理。

## 5. 关键风险

- Keyboard Extension 的 Full Access 转化与网络稳定性。
- 第三方聊天 App 输入框兼容差异。
- OCR 准确率不足导致上下文噪声。
- 用户自配 Provider 的 API Key 安全与供应商差异。
- 中英文 Prompt 质量需要真实样本回归。
- 后台管理范围扩大后，MVP 排期会比纯 iOS 输入法更长。
