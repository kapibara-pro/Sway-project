# Sway iOS

M1 iOS 主 App 最小闭环：

- 游客设备 ID：通过 `X-Sway-Device-ID` 请求头传给后端。
- Provider 配置：状态、保存、测试、清除。
- 生成页：六模式参数骨架、中文/英文、`store_allowed` / `ephemeral`、调用生成 API、展示 3 条候选、复制候选。

## 完整验证流程

下面按第一次接触 iOS 开发的视角写，目标是完成 M1 smoke：

```text
启动后端 -> 打开 iOS 工程 -> 运行 App -> 配置 mock provider -> 测试连接 -> 生成 3 条候选 -> 复制候选
```

### 1. 准备环境

需要一台 Mac，并安装：

- Xcode 15 或更新版本。
- Go 1.22 或更新版本。
- Git。

第一次打开 Xcode 时，按提示安装 iOS Simulator 和 Command Line Tools。

### 2. 拉取代码

```bash
git clone git@github.com:kapibara-pro/Sway-project.git
cd Sway-project
```

如果已经拉过仓库，更新到最新代码：

```bash
git pull --ff-only
```

确认最新提交至少包含：

```bash
git log --oneline -3
```

应该能看到类似：

```text
0f81b46 feat(ios): add M1 app provider generation flow
f8e3d0e feat(backend): add provider config and generate api
7e01662 Add product design baseline
```

### 3. 启动后端

1. 启动后端：

```bash
cd backend
go run ./cmd/server
```

看到下面日志说明后端启动成功：

```text
Sway backend listening on :8080
```

这个终端窗口不要关闭。后端默认地址是：

```text
http://127.0.0.1:8080
```

可选：再开一个新终端，检查后端健康状态：

```bash
curl http://127.0.0.1:8080/healthz
```

### 4. 打开 iOS 工程

不要打开仓库根目录，直接用 Xcode 打开工程文件：

```bash
open ios/Sway.xcodeproj
```

在 Xcode 顶部工具栏确认：

- Scheme 选择 `SwayApp`。
- 运行设备选择一个 iPhone Simulator，例如 `iPhone 15`、`iPhone 15 Pro` 或任何可用模拟器。

然后点击左上角运行按钮，或按：

```text
Command + R
```

如果 Xcode 提示自动签名或 Team 配置，模拟器运行通常不需要真实开发者账号；先选择模拟器运行即可。

### 5. 配置 mock provider

App 启动后，底部有三个 Tab：

- 生成
- 模型
- 设置

进入“模型”页，确认后端地址为：

```text
http://127.0.0.1:8080
```

然后按顺序点击：

1. `填入 Mock Provider`
2. `保存配置`
3. `测试连接`

预期结果：

- 状态显示已配置。
- Provider 显示 `mock`。
- Model 显示 `mock-chat`。
- 测试结果显示通过。

mock provider 不会调用真实 LLM，也不需要 API Key，适合 M1 smoke。

### 6. 生成 3 条候选

进入“生成”页：

1. 模式可以先选择 `回复`。
2. 语言选择 `中文`。
3. 在“上下文”里输入：

```text
你今天怎么都不找我聊天？
```

4. “我的草稿”可以先留空。
5. 语气选择 `温柔`。
6. 保存策略先选择 `保存 1 个月`。
7. 点击 `生成 3 条候选`。

预期结果：

- 页面显示“已生成 3 条候选”。
- 能看到 `request_id`。
- 列表出现 3 条候选文案。
- 点击任意候选的 `复制`，按钮变成 `已复制`。

这一步通过，就说明 iOS 主 App 到后端 mock provider 的生成闭环跑通。

### 7. 验证英文和其他模式

继续在“生成”页做一轮简单回归：

- 语言切到 `English`，再生成一次。
- 模式分别试 `改写草稿`、`破冰`、`安慰`、`道歉`、`拒绝` 中任意 1-2 个。
- 保存策略切到 `不保存原文`，再生成一次。

M1 不要求每个模式质量完美，但要求接口不报错、能返回 3 条候选。

### 8. 验证错误态

进入“模型”页，点击 `清除配置`。

再回到“生成”页点击 `生成 3 条候选`。

预期结果：

- 不应崩溃。
- 页面应显示类似“请先在 App 中配置模型”的提示。

然后回到“模型”页，重新用 mock provider 保存并测试即可恢复。

### 9. 真机验证注意事项

如果用 iPhone 真机而不是模拟器，`127.0.0.1` 指的是 iPhone 自己，不是 Mac，所以会连不上 Mac 上的后端。

真机验证时需要：

1. Mac 和 iPhone 连接同一个 Wi-Fi。
2. 在 Mac 上查看局域网 IP：

```bash
ipconfig getifaddr en0
```

通常会得到类似：

```text
192.168.1.23
```

3. App “模型”页的后端地址改成：

```text
http://192.168.1.23:8080
```

4. 确认 Mac 防火墙没有拦截 8080 端口。

### 10. M1 smoke 通过标准

以下全部通过，就可以认为 task #8 的 M1 iOS 侧 smoke 通过：

- App 能在 iPhone Simulator 或真机启动。
- “模型”页能保存 mock provider。
- “模型”页测试连接通过。
- “生成”页能成功生成 3 条候选。
- 能看到 request_id。
- 能复制候选文本。
- 清除 Provider 后，生成页能展示可理解错误，不崩溃。

## 常见问题

### Xcode 打不开工程

确认打开的是：

```text
ios/Sway.xcodeproj
```

不是仓库根目录。

### App 提示连不上后端

先确认后端终端还在运行，并看到：

```text
Sway backend listening on :8080
```

如果是模拟器，后端地址用：

```text
http://127.0.0.1:8080
```

如果是真机，使用 Mac 局域网 IP。

### 测试连接失败

先点击“填入 Mock Provider”，再点击“保存配置”，最后点击“测试连接”。只点“测试连接”但没有保存配置，会提示未配置模型。

### 复制候选后不知道是否成功

点击候选右侧“复制”后，按钮会变成“已复制”。可以打开备忘录粘贴确认。

### 需要真实模型吗

M1 smoke 不需要。先用 mock provider 跑通接口链路。真实 Provider 后续再接。

## 说明

当前环境没有 Xcode，工程文件和 Swift 源码已按 iOS 17 SwiftUI App 编写；构建验证需要在 macOS/Xcode 上执行。
